package service

import (
	"auth-service/internal/domain"
	"auth-service/internal/repository"
	"context"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// ServicioCarrito define el contrato del servicio de carritos asignados
type ServicioCarrito interface {
	ObtenerCarritosAsignados(ctx context.Context, idUsuario int32) (*domain.RespuestaCarritosAsignados, error)
	ObtenerCarritosPorCedula(ctx context.Context, cedula string) (*domain.RespuestaCarritosAsignados, error)
	ObtenerDetalladoCarrito(ctx context.Context, idUsuario int32, numCarrito int32) (*domain.RespuestaDetalladoCarrito, error)
	ObtenerDetalladoCarritoPorCedula(ctx context.Context, cedula string, numCarrito int32) (*domain.RespuestaDetalladoCarrito, error)
	ObtenerListadoPartes(ctx context.Context) (*domain.RespuestaListadoPartes, error)
	ObtenerUsuariosConCarritos(ctx context.Context) ([]domain.UsuarioConCarritos, error)
	ObtenerCarritosGenerales(ctx context.Context) (*domain.RespuestaCarritosGenerales, error)
	AsignarCarritoAUsuario(ctx context.Context, idUsuario int32, numeroCarrito int32) (*domain.RespuestaAsignacionCarrito, error)
	QuitarCarritoDeUsuario(ctx context.Context, idUsuario int32, numeroCarrito int32) (*domain.RespuestaQuitarCarrito, error)
	ObtenerUsuariosUbicacion(ctx context.Context) (*domain.RespuestaUsuariosUbicacion, error)
	ObtenerCumplimientoUsuarios(ctx context.Context) (*domain.RespuestaCumplimientoUsuarios, error)
}

type servicioCarritoImpl struct {
	repoCarrito   repository.RepositorioCarrito
	repoUbicacion repository.RepositorioUbicacion
	repoInventario repository.RepositorioInventario
	repoAdmon      repository.RepositorioAdmon
}

func NuevoServicioCarrito(rc repository.RepositorioCarrito, ru repository.RepositorioUbicacion, ri repository.RepositorioInventario, ra repository.RepositorioAdmon) ServicioCarrito {
	return &servicioCarritoImpl{
		repoCarrito:    rc,
		repoUbicacion:  ru,
		repoInventario: ri,
		repoAdmon:      ra,
	}
}

var soloDigitos = regexp.MustCompile(`\d+`)

// extraerNumeroUbicacion limpia "TECNICO 60   " → "60", "MOTO TALLER 4" → "4"
func extraerNumeroUbicacion(desc string) string {
	partes := strings.Fields(strings.TrimSpace(desc))
	for i := len(partes) - 1; i >= 0; i-- {
		if _, err := strconv.Atoi(partes[i]); err == nil {
			return partes[i]
		}
	}
	return ""
}

// extraerNumeroEmpleado limpia "98709346-\tRIVERA ZAPATA..." → "98709346"
func extraerNumeroEmpleado(empleado string) string {
	empleado = strings.TrimSpace(empleado)
	if idx := strings.Index(empleado, "-"); idx > 0 {
		return strings.TrimSpace(empleado[:idx])
	}
	return soloDigitos.FindString(empleado)
}

// ObtenerDetalladoCarrito verifica que el carrito pertenezca al usuario,
// retorna sus ítems desde SQL Server y marca cuáles ya fueron guardados
func (s *servicioCarritoImpl) ObtenerDetalladoCarrito(ctx context.Context, idUsuario int32, numCarrito int32) (*domain.RespuestaDetalladoCarrito, error) {
	// 1. Validar que el carrito pertenezca al usuario en MySQL
	numeros, err := s.repoCarrito.ObtenerCarritosPorUsuario(ctx, idUsuario)
	if err != nil {
		return nil, err
	}

	pertenece := false
	for _, n := range numeros {
		if n == numCarrito {
			pertenece = true
			break
		}
	}

	if !pertenece {
		return nil, fmt.Errorf("el carrito %d no está asignado al usuario %d", numCarrito, idUsuario)
	}

	// 2. Obtener las referencias ya guardadas en MySQL para este usuario+carrito
	referenciasGuardadas, err := s.repoInventario.ObtenerReferenciasGuardadas(ctx, idUsuario, numCarrito)
	if err != nil {
		return nil, fmt.Errorf("error al verificar completados: %w", err)
	}

	// 3. Obtener datos de SQL Server
	filas, err := s.repoUbicacion.ObtenerFilasUbicaciones(ctx)
	if err != nil {
		return nil, err
	}

	numCarritoStr := strconv.Itoa(int(numCarrito))
	idUsuarioStr := strconv.Itoa(int(idUsuario))
	var items []domain.ItemCarritoDetallado

	for _, fila := range filas {
		numUbicacion := extraerNumeroUbicacion(fila.DescUbicacion)
		numEmpleado := extraerNumeroEmpleado(fila.Empleado)

		esDelCarrito := (numUbicacion != "" && numUbicacion == numCarritoStr) ||
			(numEmpleado != "" && numEmpleado == numCarritoStr) ||
			(numEmpleado != "" && numEmpleado == idUsuarioStr && numCarritoStr == idUsuarioStr)

		if esDelCarrito {
			// Determinar completado: 1 = sí, 2 = no
			completado := 2
			refLimpia := strings.TrimSpace(fila.Referencia)
			if referenciasGuardadas[refLimpia] {
				completado = 1
			}

			items = append(items, domain.ItemCarritoDetallado{
				Referencia:  fila.Referencia,
				Ext1:        fila.Ext1,
				Descripcion: fila.Descripcion,
				UM:          fila.UM,
				Existencia:  fila.Existencia,
				Completado:  completado,
			})
		}
	}

	return &domain.RespuestaDetalladoCarrito{
		TotalProductos: len(items),
		Items:          items,
	}, nil
}

// ObtenerCarritosAsignados ejecuta el cruce MySQL ↔ SQL Server y retorna
// el detalle de registros y completados por cada carrito asignado
func (s *servicioCarritoImpl) ObtenerCarritosAsignados(ctx context.Context, idUsuario int32) (*domain.RespuestaCarritosAsignados, error) {
	// 1. Obtener carritos del usuario desde MySQL
	numerosCarrito, err := s.repoCarrito.ObtenerCarritosPorUsuario(ctx, idUsuario)
	if err != nil {
		return nil, fmt.Errorf("error al obtener carritos MySQL: %w", err)
	}

	// 2. Obtener completados por carrito desde registros_inventario
	completadosPorCarrito, err := s.repoInventario.ContarCompletadosPorUsuario(ctx, idUsuario)
	if err != nil {
		return nil, fmt.Errorf("error al contar completados: %w", err)
	}

	// 3. Obtener todas las filas de vw_Ubicaciones en SQL Server
	filas, err := s.repoUbicacion.ObtenerFilasUbicaciones(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al obtener ubicaciones SQL Server: %w", err)
	}

	idUsuarioStr := strconv.Itoa(int(idUsuario))
	detalles := make([]domain.DetalleCarrito, 0, len(numerosCarrito))
	setCarritos := make(map[string]int, len(numerosCarrito))

	for i, num := range numerosCarrito {
		numStr := strconv.Itoa(int(num))
		setCarritos[numStr] = i
		detalles = append(detalles, domain.DetalleCarrito{
			NumeroCarrito: num,
			Registros:     0,
			Completados:   completadosPorCarrito[num], // conteo desde MySQL
		})
	}

	totalRegistrosGeneral := 0

	// 4. Procesar filas de SQL Server
	for _, fila := range filas {
		numUbicacion := extraerNumeroUbicacion(fila.DescUbicacion)
		numEmpleado := extraerNumeroEmpleado(fila.Empleado)

		encontradoParaUsuario := false

		// Función auxiliar para extraer el nombre del empleado de forma limpia
		extraerNombre := func(s string) string {
			if idx := strings.Index(s, "-"); idx >= 0 {
				return strings.TrimSpace(s[idx+1:])
			}
			return ""
		}

		if numUbicacion != "" {
			if idx, ok := setCarritos[numUbicacion]; ok {
				detalles[idx].Registros++
				encontradoParaUsuario = true
				// Capturar el nombre real desde la columna Empleado (después del guion)
				if detalles[idx].Nombre == "" {
					detalles[idx].Nombre = extraerNombre(fila.Empleado)
				}
			}
		}

		if numEmpleado != "" {
			if idx, ok := setCarritos[numEmpleado]; ok {
				detalles[idx].Registros++
				encontradoParaUsuario = true
				// Capturar el nombre real desde la columna Empleado también aquí
				if detalles[idx].Nombre == "" {
					detalles[idx].Nombre = extraerNombre(fila.Empleado)
				}
			}
		}

		if numEmpleado != "" && numEmpleado == idUsuarioStr {
			encontradoParaUsuario = true
		}

		if encontradoParaUsuario {
			totalRegistrosGeneral++
		}
	}

	return &domain.RespuestaCarritosAsignados{
		IDUsuario:      idUsuario,
		Carritos:       detalles,
		TotalRegistros: totalRegistrosGeneral,
	}, nil
}

func (s *servicioCarritoImpl) ObtenerListadoPartes(ctx context.Context) (*domain.RespuestaListadoPartes, error) {
	marcas, err := s.repoAdmon.ObtenerMarcas(ctx)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo marcas: %w", err)
	}

	total, err := s.repoAdmon.ContarTotalPartes(ctx)
	if err != nil {
		return nil, fmt.Errorf("error contando partes: %w", err)
	}

	return &domain.RespuestaListadoPartes{
		TotalHerramientas: total,
		Marcas:            marcas,
	}, nil
}

func (s *servicioCarritoImpl) ObtenerUsuariosConCarritos(ctx context.Context) ([]domain.UsuarioConCarritos, error) {
	usuarios, err := s.repoCarrito.ObtenerUsuariosConCarritos(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al obtener usuarios con carritos: %w", err)
	}

	return usuarios, nil
}

func (s *servicioCarritoImpl) ObtenerCarritosGenerales(ctx context.Context) (*domain.RespuestaCarritosGenerales, error) {
	filas, err := s.repoUbicacion.ObtenerFilasUbicaciones(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al obtener filas de ubicaciones: %w", err)
	}

	porNumero := make(map[string]domain.CarritoGeneral)

	for _, fila := range filas {
		numeroTexto := extraerNumeroUbicacion(fila.DescUbicacion)
		nombreCompleto := strings.TrimSpace(fila.DescUbicacion)
		cedula := extraerNumeroEmpleado(fila.Empleado)

		numeroCarrito := numeroTexto
		if numeroCarrito == "" {
			numeroCarrito = cedula
		}

		if numeroCarrito == "" {
			continue
		}

		actual, existe := porNumero[numeroCarrito]
		if !existe {
			porNumero[numeroCarrito] = domain.CarritoGeneral{
				NumeroCarrito: numeroCarrito,
				Cedula:        cedula,
				NombreCompleto: nombreCompleto,
			}
			continue
		}

		if actual.Cedula == "" && cedula != "" {
			actual.Cedula = cedula
		}
		if actual.NombreCompleto == "" && nombreCompleto != "" {
			actual.NombreCompleto = nombreCompleto
		}

		porNumero[numeroCarrito] = actual
	}

	carritos := make([]domain.CarritoGeneral, 0, len(porNumero))
	for _, carrito := range porNumero {
		carritos = append(carritos, carrito)
	}

	sort.Slice(carritos, func(i, j int) bool {
		numI, errI := strconv.Atoi(carritos[i].NumeroCarrito)
		numJ, errJ := strconv.Atoi(carritos[j].NumeroCarrito)

		if errI == nil && errJ == nil {
			return numI < numJ
		}

		if errI == nil {
			return true
		}

		if errJ == nil {
			return false
		}

		return carritos[i].NumeroCarrito < carritos[j].NumeroCarrito
	})

	return &domain.RespuestaCarritosGenerales{
		TotalCarritos: len(carritos),
		Carritos:      carritos,
	}, nil
}

func (s *servicioCarritoImpl) AsignarCarritoAUsuario(ctx context.Context, idUsuario int32, numeroCarrito int32) (*domain.RespuestaAsignacionCarrito, error) {
	existe, err := s.repoCarrito.ExisteUsuarioOperarioActivo(ctx, idUsuario)
	if err != nil {
		return nil, fmt.Errorf("error validando usuario destino: %w", err)
	}
	if !existe {
		return nil, fmt.Errorf("el usuario %d no existe o no es operario activo", idUsuario)
	}

	usuariosPrevios, err := s.repoCarrito.AsignarCarritoAUsuario(ctx, idUsuario, numeroCarrito)
	if err != nil {
		return nil, err
	}

	transferido := false
	for _, idPrevio := range usuariosPrevios {
		if idPrevio != idUsuario {
			transferido = true
			break
		}
	}

	mensaje := "carrito asignado correctamente"
	if transferido {
		mensaje = "carrito transferido correctamente"
	}

	return &domain.RespuestaAsignacionCarrito{
		IDUsuarioDestino: idUsuario,
		NumeroCarrito:    numeroCarrito,
		Transferido:      transferido,
		UsuariosPrevios:  usuariosPrevios,
		Mensaje:          mensaje,
	}, nil
}

func (s *servicioCarritoImpl) QuitarCarritoDeUsuario(ctx context.Context, idUsuario int32, numeroCarrito int32) (*domain.RespuestaQuitarCarrito, error) {
	quitado, err := s.repoCarrito.QuitarCarritoDeUsuario(ctx, idUsuario, numeroCarrito)
	if err != nil {
		return nil, err
	}

	mensaje := "carrito quitado correctamente"
	if !quitado {
		mensaje = "el carrito no estaba asignado a ese usuario"
	}

	return &domain.RespuestaQuitarCarrito{
		IDUsuario:     idUsuario,
		NumeroCarrito: numeroCarrito,
		Quitado:       quitado,
		Mensaje:       mensaje,
	}, nil
}

// ObtenerUsuariosUbicacion retorna los usuarios únicos de vw_Ubicaciones
func (s *servicioCarritoImpl) ObtenerUsuariosUbicacion(ctx context.Context) (*domain.RespuestaUsuariosUbicacion, error) {
	usuarios, err := s.repoUbicacion.ObtenerUsuariosUnicos(ctx)
	if err != nil {
		return nil, err
	}

	return &domain.RespuestaUsuariosUbicacion{
		Total:    len(usuarios),
		Usuarios: usuarios,
	}, nil
}

// ObtenerDetalladoCarritoPorCedula retorna los ítems de un carrito específico buscando por cédula
// Sin validación de MySQL (usado por visualizadores)
func (s *servicioCarritoImpl) ObtenerDetalladoCarritoPorCedula(ctx context.Context, cedula string, numCarrito int32) (*domain.RespuestaDetalladoCarrito, error) {
	// 1. Obtener todas las filas de SQL Server
	filas, err := s.repoUbicacion.ObtenerFilasUbicaciones(ctx)
	if err != nil {
		return nil, err
	}

	numCarritoStr := strconv.Itoa(int(numCarrito))
	var items []domain.ItemCarritoDetallado

	for _, fila := range filas {
		numUbicacion := extraerNumeroUbicacion(fila.DescUbicacion)
		numEmpleado := extraerNumeroEmpleado(fila.Empleado)

		// El carrito debe coincidir con el número solicitado
		esDelCarrito := (numUbicacion != "" && numUbicacion == numCarritoStr) ||
			(numEmpleado != "" && numEmpleado == numCarritoStr)

		// Y debe pertenecer al empleado (por cédula)
		perteneceAEmpleado := numEmpleado == cedula

		if esDelCarrito && perteneceAEmpleado {
			items = append(items, domain.ItemCarritoDetallado{
				Referencia:  fila.Referencia,
				Ext1:        fila.Ext1,
				Descripcion: fila.Descripcion,
				UM:          fila.UM,
				Existencia:  fila.Existencia,
				Completado:  2, // No completado por defecto para visualizadores
			})
		}
	}

	return &domain.RespuestaDetalladoCarrito{
		TotalProductos: len(items),
		Items:          items,
	}, nil
}

// ObtenerCarritosPorCedula busca carritos en vw_Ubicaciones donde el empleado coincida con la cédula
// Usado por visualizadores (empleados que no están en la tabla usuarios de MySQL)
func (s *servicioCarritoImpl) ObtenerCarritosPorCedula(ctx context.Context, cedula string) (*domain.RespuestaCarritosAsignados, error) {
	// 1. Obtener todas las filas de vw_Ubicaciones
	filas, err := s.repoUbicacion.ObtenerFilasUbicaciones(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al obtener filas de ubicaciones: %w", err)
	}

	// 2. Buscar carritos donde el empleado coincida con la cédula
	carritosMap := make(map[string]*domain.DetalleCarrito)

	for _, fila := range filas {
		numEmpleado := extraerNumeroEmpleado(fila.Empleado)
		if numEmpleado == cedula {
			// Encontrar el número del carrito desde DescUbicacion
			numCarrito := extraerNumeroUbicacion(fila.DescUbicacion)
			if numCarrito == "" {
				// Si no hay número en ubicación, usar la cédula como identificador
				numCarrito = cedula
			}

			if carrito, existe := carritosMap[numCarrito]; existe {
				carrito.Registros++
			} else {
				carritosMap[numCarrito] = &domain.DetalleCarrito{
					NumeroCarrito: parsearInt32(numCarrito),
					Nombre:        strings.TrimSpace(fila.DescUbicacion),
					Registros:     1,
					Completados:   0,
				}
			}
		}
	}

	// 3. Convertir map a slice
	detalles := make([]domain.DetalleCarrito, 0, len(carritosMap))
	for _, carrito := range carritosMap {
		detalles = append(detalles, *carrito)
	}

	// Ordenar por número de carrito
	sort.Slice(detalles, func(i, j int) bool {
		return detalles[i].NumeroCarrito < detalles[j].NumeroCarrito
	})

	return &domain.RespuestaCarritosAsignados{
		IDUsuario:      -1, // Visualizador no tiene ID real
		Carritos:       detalles,
		TotalRegistros: len(filas),
	}, nil
}

// parsearInt32 convierte string a int32, retorna 0 si hay error
func parsearInt32(s string) int32 {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return int32(n)
}

// ObtenerCumplimientoUsuarios retorna el cumplimiento diario de todos los usuarios con carritos asignados
func (s *servicioCarritoImpl) ObtenerCumplimientoUsuarios(ctx context.Context) (*domain.RespuestaCumplimientoUsuarios, error) {
	// 1. Obtener todos los usuarios con carritos
	usuarios, err := s.repoCarrito.ObtenerUsuariosConCarritos(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al obtener usuarios con carritos: %w", err)
	}

	// 2. Obtener todas las filas de ubicaciones para contar registros por carrito
	filas, err := s.repoUbicacion.ObtenerFilasUbicaciones(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al obtener ubicaciones: %w", err)
	}

	// 3. Construir mapa de conteo de registros por número de carrito
	registrosPorCarrito := make(map[string]int)
	for _, fila := range filas {
		numUbicacion := extraerNumeroUbicacion(fila.DescUbicacion)
		numEmpleado := extraerNumeroEmpleado(fila.Empleado)

		if numUbicacion != "" {
			registrosPorCarrito[numUbicacion]++
		}
		if numEmpleado != "" && numEmpleado != numUbicacion {
			registrosPorCarrito[numEmpleado]++
		}
	}

	// 4. Procesar cada usuario
	cumplimientos := make([]domain.CumplimientoUsuario, 0, len(usuarios))

	for _, usuario := range usuarios {
		// Obtener completados de hoy para este usuario
		completadosPorCarrito, err := s.repoInventario.ContarCompletadosPorUsuario(ctx, usuario.IDUsuario)
		if err != nil {
			continue // Si falla, continuamos con el siguiente usuario
		}

		// Construir cumplimiento por carrito
		carritosCumplimiento := make([]domain.CumplimientoCarrito, 0, len(usuario.Carritos))
		totalRegistros := 0
		totalCompletados := 0

		for _, numCarrito := range usuario.Carritos {
			numCarritoStr := strconv.Itoa(int(numCarrito))
			registros := registrosPorCarrito[numCarritoStr]
			completados := completadosPorCarrito[numCarrito]
			pendientes := registros - completados
			if pendientes < 0 {
				pendientes = 0
			}

			var porcentaje float64
			if registros > 0 {
				porcentaje = float64(completados) * 100.0 / float64(registros)
			}

			carritosCumplimiento = append(carritosCumplimiento, domain.CumplimientoCarrito{
				NumeroCarrito: numCarrito,
				Registros:     registros,
				Completados:   completados,
				Pendientes:    pendientes,
				Porcentaje:    math.Round(porcentaje*10) / 10, // Redondear a 1 decimal
			})

			totalRegistros += registros
			totalCompletados += completados
		}

		totalPendientes := totalRegistros - totalCompletados
		if totalPendientes < 0 {
			totalPendientes = 0
		}

		var porcentajeGlobal float64
		if totalRegistros > 0 {
			porcentajeGlobal = float64(totalCompletados) * 100.0 / float64(totalRegistros)
		}

		// Determinar estado global
		estado := "pendiente"
		if totalRegistros == 0 {
			estado = "sin_registros"
		} else if porcentajeGlobal >= 100 {
			estado = "completado"
		} else if porcentajeGlobal > 0 {
			estado = "en_progreso"
		}

		cumplimientos = append(cumplimientos, domain.CumplimientoUsuario{
			IDUsuario:        usuario.IDUsuario,
			Nombre:           usuario.Nombre,
			Correo:           usuario.Correo,
			Carritos:         carritosCumplimiento,
			TotalRegistros:   totalRegistros,
			TotalCompletados: totalCompletados,
			TotalPendientes:  totalPendientes,
			PorcentajeGlobal: math.Round(porcentajeGlobal*10) / 10,
			Estado:           estado,
		})
	}

	return &domain.RespuestaCumplimientoUsuarios{
		Total:    len(cumplimientos),
		Usuarios: cumplimientos,
	}, nil
}
