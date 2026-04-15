-- name: ObtenerUsuarioPorEmail :one
SELECT
  id_usuario,
  empleado,
  nombre_completo,
  descripcion_cargo,
  email,
  contrasena,
  rol,
  activo,
  fecha_creacion,
  fecha_actualizacion
FROM usuarios
WHERE email = ? AND activo = 1
LIMIT 1;

-- name: ObtenerUsuarioPorId :one
SELECT
  id_usuario,
  empleado,
  nombre_completo,
  descripcion_cargo,
  email,
  contrasena,
  rol,
  activo,
  fecha_creacion,
  fecha_actualizacion
FROM usuarios
WHERE id_usuario = ? AND activo = 1
LIMIT 1;
