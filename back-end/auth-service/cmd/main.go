package main

import (
	"auth-service/internal/config"
	"auth-service/internal/db/sqlc"
	"auth-service/internal/handlers"
	"auth-service/internal/middleware"
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/microsoft/go-mssqldb"
)

func main() {
	// 1. Cargar configuración
	cfg, err := config.CargarConfiguracion()
	if err != nil {
		log.Fatalf("error al cargar configuración: %v", err)
	}

	// 2. Conectar a MySQL
	dbMySQL, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		log.Fatalf("error al abrir conexión con MySQL: %v", err)
	}
	defer dbMySQL.Close()

	if err := dbMySQL.Ping(); err != nil {
		log.Fatalf("no se puede conectar a MySQL: %v", err)
	}
	log.Println("✅ Conectado a MySQL correctamente")

	// 3. Conectar a SQL Server (Reportes)
	dbSQLServer, err := sql.Open("sqlserver", cfg.DSNSQLServer())
	if err != nil {
		log.Fatalf("error al abrir conexión con SQL Server: %v", err)
	}
	defer dbSQLServer.Close()

	if err := dbSQLServer.Ping(); err != nil {
		log.Printf("⚠️  No se puede conectar a SQL Server (el servicio de carritos no estará disponible): %v", err)
	} else {
		log.Println("✅ Conectado a SQL Server correctamente")
	}

	// 3b. Conectar a SQL Server UNOEE (para préstamo)
	dbUNOEE, err := sql.Open("sqlserver", cfg.DSNSQLServerUNOEE())
	if err != nil {
		log.Printf("⚠️  Error al abrir conexión con SQL Server UNOEE: %v", err)
	}
	defer dbUNOEE.Close()

	if err := dbUNOEE.Ping(); err != nil {
		log.Printf("⚠️  No se puede conectar a SQL Server UNOEE (el servicio de préstamo no estará disponible): %v", err)
	} else {
		log.Println("✅ Conectado a SQL Server UNOEE correctamente")
	}

	// 4. Conectar a MySQL ADMON
	dbAdmon, err := sql.Open("mysql", cfg.DSNAdmon())
	if err != nil {
		log.Printf("⚠️  Error al abrir conexión con MySQL ADMON: %v", err)
	}
	defer dbAdmon.Close()

	if err := dbAdmon.Ping(); err != nil {
		log.Printf("⚠️  No se puede conectar a MySQL ADMON (el listado de partes no estará disponible): %v", err)
	} else {
		log.Println("✅ Conectado a MySQL ADMON correctamente")
	}

	// 4. Inyección de dependencias - Repositorios base
	queries := sqlc.New(dbMySQL)
	repoUsuario := repository.NuevoRepositorioUsuario(dbMySQL, queries)
	repoCarrito := repository.NuevoRepositorioCarrito(dbMySQL)
	repoUbicacion := repository.NuevoRepositorioUbicacion(dbSQLServer)

	// 5. Inyección de dependencias - Auth (con ubicaciones para login por cédula)
	svcAuth := service.NuevoServicioAuthConUbicacion(repoUsuario, repoUbicacion, cfg.JWTSecreto, cfg.JWTExpiracionHoras)
	manejadorAuth := handlers.NuevoManejadorAuth(svcAuth)
	repoAdmon := repository.NuevoRepositorioAdmon(dbAdmon)
	repoUNOEE := repository.NuevoRepositorioUNOEE(dbUNOEE)

	// 6. Inyección de dependencias - Inventario
	repoInventario := repository.NuevoRepositorioInventario(dbMySQL)
	svcInventario := service.NuevoServicioInventario(repoInventario)
	manejadorInventario := handlers.NuevoManejadorInventario(svcInventario)

	// El servicio de carritos también necesita repoInventario para saber qué está completado
	svcCarrito := service.NuevoServicioCarrito(repoCarrito, repoUbicacion, repoInventario, repoAdmon)
	manejadorCarritos := handlers.NuevoManejadorCarritos(svcCarrito)

	// Servicio de préstamo
	svcPrestamo := service.NuevoServicioPrestamo(repoUNOEE)
	manejadorPrestamo := handlers.NuevoManejadorPrestamo(svcPrestamo)

	// Servicio de préstamo CRUD (gestión de préstamos a operarios)
	repoPrestamo := repository.NuevoRepositorioPrestamo(dbMySQL)
	svcPrestamoCRUD := service.NuevoServicioPrestamoCRUD(repoPrestamo, repoUNOEE)
	manejadorPrestamoCRUD := handlers.NuevoManejadorPrestamoCRUD(svcPrestamoCRUD)

	// 7. Configurar rutas con prefijo /api
	mux := http.NewServeMux()

	// Rutas públicas de auth
	mux.HandleFunc("/api/auth/login", manejadorAuth.Login)
	mux.HandleFunc("/api/auth/login-cedula", manejadorAuth.LoginPorCedula)
	mux.HandleFunc("/api/auth/verificar", manejadorAuth.VerificarToken)

	// Ruta de perfil (protegida por JWT)
	mux.Handle("/api/auth/perfil", middleware.AutenticarJWT(svcAuth)(
		http.HandlerFunc(manejadorAuth.Perfil),
	))

	// Ruta para crear usuarios (solo admin, protegida por JWT)
	mux.Handle("/api/auth/usuarios", middleware.AutenticarJWT(svcAuth)(
		http.HandlerFunc(manejadorAuth.CrearUsuario),
	))

	// Ruta de carritos asignados (protegida por JWT)
	// GET /api/carritos/asignados/{id_usuario}
	mux.Handle("/api/carritos/asignados", middleware.AutenticarJWT(svcAuth)(
		http.HandlerFunc(manejadorCarritos.CarritosAsignados),
	))

	// Ruta de detalle de carrito (protegida por JWT)
	// GET /api/carritos/detallado/{id_usuario}/{numero_carrito}
	mux.Handle("/api/carritos/detallado", middleware.AutenticarJWT(svcAuth)(
		http.HandlerFunc(manejadorCarritos.DetalladoCarrito),
	))

	// Ruta para guardar inventario (protegida por JWT)
	// POST /api/inventario/guardar
	mux.Handle("/api/inventario/guardar", middleware.AutenticarJWT(svcAuth)(
		http.HandlerFunc(manejadorInventario.GuardarInventario),
	))

	// Ruta para obtener listado de marcas protegida por JWT
	// GET /api/carritos/listado-partes
	mux.Handle("/api/carritos/listado-partes", middleware.AutenticarJWT(svcAuth)(
		http.HandlerFunc(manejadorCarritos.ListadoPartes),
	))

	// Ruta para listar usuarios con sus carritos (protegida por JWT)
	// GET /api/carritos/usuarios
	mux.Handle("/api/carritos/usuarios", middleware.AutenticarJWT(svcAuth)(
		http.HandlerFunc(manejadorCarritos.UsuariosConCarritos),
	))

	// Ruta para asignar/transferir carrito (protegida por JWT)
	// POST /api/carritos/asignar
	mux.Handle("/api/carritos/asignar", middleware.AutenticarJWT(svcAuth)(
		http.HandlerFunc(manejadorCarritos.AsignarCarrito),
	))

	// Ruta para quitar carrito de un usuario (protegida por JWT)
	// POST /api/carritos/quitar
	mux.Handle("/api/carritos/quitar", middleware.AutenticarJWT(svcAuth)(
		http.HandlerFunc(manejadorCarritos.QuitarCarrito),
	))

	// Ruta para listar todos los carritos existentes (protegida por JWT)
	// GET /api/carritos/general
	mux.Handle("/api/carritos/general", middleware.AutenticarJWT(svcAuth)(
		http.HandlerFunc(manejadorCarritos.CarritosGenerales),
	))

	// Ruta para ver cumplimiento diario de todos los usuarios (protegida por JWT)
	// GET /api/carritos/cumplimiento
	mux.Handle("/api/carritos/cumplimiento", middleware.AutenticarJWT(svcAuth)(
		http.HandlerFunc(manejadorCarritos.CumplimientoUsuarios),
	))

	// Ruta para listar usuarios únicos de vw_Ubicaciones (protegida por JWT)
	// GET /api/carritos/usuarios-ubicacion
	mux.Handle("/api/carritos/usuarios-ubicacion", middleware.AutenticarJWT(svcAuth)(
		http.HandlerFunc(manejadorCarritos.UsuariosUbicacion),
	))

	// Ruta para listar ítems disponibles para préstamo (protegida por JWT)
	// GET /api/prestamo/items
	mux.Handle("/api/prestamo/items", middleware.AutenticarJWT(svcAuth)(
		http.HandlerFunc(manejadorPrestamo.ItemsPrestamo),
	))

	// Rutas CRUD de préstamos a operarios (protegidas por JWT)
	// POST /api/prestamos - Crear préstamo
	mux.Handle("/api/prestamos", middleware.AutenticarJWT(svcAuth)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				manejadorPrestamoCRUD.CrearPrestamo(w, r)
			} else if r.Method == http.MethodGet {
				manejadorPrestamoCRUD.ListarPrestamos(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		}),
	))

	// GET /api/prestamos/ - Obtener o listar (maneja /prestamos/:id)
	mux.Handle("/api/prestamos/", middleware.AutenticarJWT(svcAuth)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			// Si contiene /devolver es devolución
			if strings.Contains(path, "/devolver") {
				manejadorPrestamoCRUD.DevolverPrestamo(w, r)
			} else if strings.Contains(path, "/operario/") {
				manejadorPrestamoCRUD.ListarPrestamosPorOperario(w, r)
			} else {
				// Es obtener por ID: /prestamos/:id
				manejadorPrestamoCRUD.ObtenerPrestamo(w, r)
			}
		}),
	))

	// 7. Iniciar servidor con CORS
	direccion := fmt.Sprintf(":%s", cfg.Puerto)
	log.Printf("🚀 Auth Service corriendo en http://localhost%s", direccion)
	log.Fatal(http.ListenAndServe(direccion, aplicarCORS(mux)))
}

// aplicarCORS agrega headers necesarios para el frontend
func aplicarCORS(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		siguiente.ServeHTTP(w, r)
	})
}
