-- Migración: creación de la tabla prestamos
CREATE TABLE IF NOT EXISTS `prestamos` (
  `id_prestamo` bigint NOT NULL AUTO_INCREMENT,

  -- Identificación del ítem prestado (de UNOEE)
  `referencia` varchar(50) NOT NULL COMMENT 'f_referencia de ADMIN-INVENTARIO_TODAS_BODEGAS',
  `descripcion` varchar(255) DEFAULT NULL COMMENT 'f_desc_item para referencia rápida',
  `ext1` varchar(50) DEFAULT NULL COMMENT 'f121_id_ext1_detalle (marca)',
  `um` varchar(10) DEFAULT NULL COMMENT 'f_um (unidad de medida)',

  -- Persona que recibe el préstamo (de vw_Ubicaciones)
  `cedula_operario` varchar(20) NOT NULL COMMENT 'Cédula del empleado que recibe',
  `nombre_operario` varchar(255) DEFAULT NULL COMMENT 'Nombre completo del operario',

  -- Cantidad prestada
  `cantidad_prestada` decimal(12,2) NOT NULL DEFAULT 1 COMMENT 'Cuántas unidades se prestaron',

  -- Estados del préstamo
  `estado` enum('activo','devuelto','parcial') NOT NULL DEFAULT 'activo'
    COMMENT 'activo: en préstamo, devuelto: todo devuelto, parcial: parte devuelta',

  -- Fechas
  `fecha_prestamo` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `fecha_devolucion` timestamp NULL DEFAULT NULL COMMENT 'Cuando se devolvió completamente',

  -- Auditoría
  `id_usuario_prestamista` int DEFAULT NULL COMMENT 'ID del usuario del sistema que hizo el préstamo',
  `nombre_usuario_prestamista` varchar(255) DEFAULT NULL COMMENT 'Nombre del usuario que registró',

  -- Metadatos
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (`id_prestamo`),
  KEY `idx_referencia` (`referencia`),
  KEY `idx_cedula_operario` (`cedula_operario`),
  KEY `idx_estado` (`estado`),
  KEY `idx_fecha_prestamo` (`fecha_prestamo`),
  KEY `idx_activos` (`referencia`, `estado`) COMMENT 'Para consultar stock disponible'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
