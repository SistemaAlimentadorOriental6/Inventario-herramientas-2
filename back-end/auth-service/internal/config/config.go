package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Configuracion contiene todos los parámetros del microservicio
type Configuracion struct {
	Puerto             string
	DBHost             string
	DBPuerto           string
	DBUsuario          string
	DBContrasena       string
	DBNombre           string
	JWTSecreto         string
	JWTExpiracionHoras int

	// SQL Server - BD de Reportes
	SQLServerUsuario            string
	SQLServerContrasena         string
	SQLServerHost               string
	SQLServerBaseDatos          string
	SQLServerEncriptar          string
	SQLServerConfiarCertificado string

	// SQL Server - BD UNOEE (para préstamo)
	SQLServerUNOEEBaseDatos string

	// MySQL ADMON
	AdmonHost     string
	AdmonPuerto   string
	AdmonUsuario  string
	AdmonContrasena string
	AdmonNombre   string
}

// CargarConfiguracion lee variables de entorno desde .env
func CargarConfiguracion() (*Configuracion, error) {
	_ = godotenv.Load()

	expiracion, err := strconv.Atoi(obtenerEnv("JWT_EXPIRACION_HORAS", "24"))
	if err != nil {
		return nil, fmt.Errorf("JWT_EXPIRACION_HORAS inválido: %w", err)
	}

	return &Configuracion{
		Puerto:             obtenerEnv("PUERTO", "8080"),
		DBHost:             obtenerEnv("DB_HOST", "localhost"),
		DBPuerto:           obtenerEnv("DB_PORT", "3306"),
		DBUsuario:          obtenerEnv("DB_USUARIO", "root"),
		DBContrasena:       obtenerEnv("DB_CONTRASENA", ""),
		DBNombre:           obtenerEnv("DB_NOMBRE", "bdsaocomco_inventario_carrito"),
		JWTSecreto:         obtenerEnv("JWT_SECRETO", "secreto_por_defecto"),
		JWTExpiracionHoras: expiracion,

		SQLServerUsuario:            obtenerEnv("DB_USER", ""),
		SQLServerContrasena:         obtenerEnv("DB_PASSWORD", ""),
		SQLServerHost:               obtenerEnv("DB_SERVER", ""),
		SQLServerBaseDatos:          obtenerEnv("DB_DATABASE", "Reportes"),
		SQLServerEncriptar:          obtenerEnv("DB_ENCRYPT", "false"),
		SQLServerConfiarCertificado: obtenerEnv("DB_TRUST_CERTIFICATE", "true"),
		SQLServerUNOEEBaseDatos:     obtenerEnv("DB_DATABASE2", "UNOEE"),

		AdmonHost:     obtenerEnv("MYSQL_ADMON_HOST", "localhost"),
		AdmonPuerto:   obtenerEnv("MYSQL_ADMON_PORT", "3306"),
		AdmonUsuario:  obtenerEnv("MYSQL_ADMON_USER", ""),
		AdmonContrasena: obtenerEnv("MYSQL_ADMON_PASSWORD", ""),
		AdmonNombre:   obtenerEnv("MYSQL_ADMON_DATABASE", ""),
	}, nil
}

// DSN genera la cadena de conexión para MySQL
func (c *Configuracion) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=America%%2FBogota",
		c.DBUsuario,
		c.DBContrasena,
		c.DBHost,
		c.DBPuerto,
		c.DBNombre,
	)
}
// DSNAdmon genera la cadena de conexión para la base de datos ADMON (MySQL)
func (c *Configuracion) DSNAdmon() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=America%%2FBogota",
		c.AdmonUsuario,
		c.AdmonContrasena,
		c.AdmonHost,
		c.AdmonPuerto,
		c.AdmonNombre,
	)
}

// DSNSQLServer genera la cadena de conexión para SQL Server
func (c *Configuracion) DSNSQLServer() string {
	return fmt.Sprintf(
		"sqlserver://%s:%s@%s?database=%s&encrypt=%s&TrustServerCertificate=%s",
		c.SQLServerUsuario,
		c.SQLServerContrasena,
		c.SQLServerHost,
		c.SQLServerBaseDatos,
		c.SQLServerEncriptar,
		c.SQLServerConfiarCertificado,
	)
}

// DSNSQLServerUNOEE genera la cadena de conexión para SQL Server UNOEE
func (c *Configuracion) DSNSQLServerUNOEE() string {
	return fmt.Sprintf(
		"sqlserver://%s:%s@%s?database=%s&encrypt=%s&TrustServerCertificate=%s",
		c.SQLServerUsuario,
		c.SQLServerContrasena,
		c.SQLServerHost,
		c.SQLServerUNOEEBaseDatos,
		c.SQLServerEncriptar,
		c.SQLServerConfiarCertificado,
	)
}

func obtenerEnv(clave, porDefecto string) string {
	if valor := os.Getenv(clave); valor != "" {
		return valor
	}
	return porDefecto
}
