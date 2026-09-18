-- -----------------------------------------------------------
-- 1. Habilitar la extensión PostGIS
-- -----------------------------------------------------------
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS postgis_topology;


-- -----------------------------------------------------------
-- 2. Tablas de Dimensiones y Usuarios
-- -----------------------------------------------------------

-- Tabla: genero
CREATE TABLE genero (
    id SERIAL PRIMARY KEY,
    tipo VARCHAR(20) NOT NULL UNIQUE,
    descripcion VARCHAR(255)
);

-- Tabla: roles
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL UNIQUE
);

-- Tabla: usuario
CREATE TABLE usuario (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(100) UNIQUE NOT NULL,
    apepat VARCHAR(100) NOT NULL,
    apemat VARCHAR(100),
    foto BYTEA,
    correo VARCHAR(50) UNIQUE,
    password VARCHAR(70),
    fecharegistro TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    fechaact TIMESTAMP WITH TIME ZONE,
    idrol INTEGER REFERENCES roles(id),
    idgenero INTEGER REFERENCES genero(id),
    esactivo BOOLEAN DEFAULT TRUE
);

-- Tabla: comentarios (Relacionada con usuario)
CREATE TABLE comentarios (
    id SERIAL PRIMARY KEY,
    descripcion VARCHAR(300) NOT NULL,
    fecharegistro TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    idusuario INTEGER REFERENCES usuario(id) ON DELETE CASCADE,
    contenido TEXT
);

-- Tabla: notificaciones (Relacionada con usuario)
CREATE TABLE notificaciones (
    id SERIAL PRIMARY KEY,
    cantidad INTEGER DEFAULT 1,
    tonos BOOLEAN,
    push BOOLEAN,
    fecharegistro TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    fechaact TIMESTAMP WITH TIME ZONE,
    idusuario INTEGER REFERENCES usuario(id) ON DELETE CASCADE,
    contenido TEXT
);

-- Tabla: historial (Relacionada con usuario)
CREATE TABLE historial (
    id SERIAL PRIMARY KEY,
    fecha TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    rutas JSONB,
    idusuario INTEGER REFERENCES usuario(id) ON DELETE CASCADE,
    detalle TEXT
);

-- Tabla: direcciones (Relacionada con usuario)
CREATE TABLE direcciones (
    id SERIAL PRIMARY KEY,
    alias VARCHAR(300),
    rutas JSONB,
    fecharegistro TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    fechaact TIMESTAMP WITH TIME ZONE,
    idusuario INTEGER REFERENCES usuario(id) ON DELETE CASCADE
);


-- -----------------------------------------------------------
-- 3. Tablas de Entidades Geográficas y Relacionales
-- -----------------------------------------------------------

-- Tabla: estadion (Estadio/Estación)
CREATE TABLE estadion (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(100) UNIQUE NOT NULL,
    latitud DOUBLE PRECISION NOT NULL,
    longitud DOUBLE PRECISION NOT NULL,
    urlweb VARCHAR(300),
    descripcion TEXT,
    numero INTEGER,
    altitud DOUBLE PRECISION,
    habilidad VARCHAR(50),
    -- Columna espacial para el punto de la estación
    geom GEOMETRY(Point, 4326)
);
-- Índice espacial para búsquedas rápidas
CREATE INDEX idx_estadion_geom ON estadion USING GIST (geom);


-- Tabla: linea (Línea o Ruta)
CREATE TABLE linea (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(50) UNIQUE NOT NULL,
    numero INTEGER,
    descripcion VARCHAR(200),
    color VARCHAR(20),
    imagen TEXT,
    habilidad VARCHAR(50)
);

-- Tabla: estacioneslinea (Relaciona estación con línea y contiene su geometría)
CREATE TABLE estacionlinea (
    id SERIAL PRIMARY KEY,
    idestacion INTEGER REFERENCES estadion(id) ON DELETE CASCADE,
    idlinea INTEGER REFERENCES linea(id) ON DELETE CASCADE,
    numero INTEGER,
    -- Atributos de la geometría (se pueden usar para el punto de conexión)
    longitud DOUBLE PRECISION,
    latitud DOUBLE PRECISION,
    altitud DOUBLE PRECISION,
    habilidad VARCHAR(50),
    -- Columna espacial: El punto de la conexión (Point)
    geom GEOMETRY(Point, 4326)
);
-- Índice espacial
CREATE INDEX idx_estacionlinea_geom ON estacionlinea USING GIST (geom);


-- Tabla: distancias (Calcula la distancia entre estaciones y líneas)
CREATE TABLE distancias (
    id SERIAL PRIMARY KEY,
    idestacionb INTEGER REFERENCES estadion(id) ON DELETE CASCADE,
    idlinea INTEGER REFERENCES linea(id) ON DELETE CASCADE,
    -- Atributos de la distancia
    longitud DOUBLE PRECISION,
    distancia DOUBLE PRECISION,
    -- Columna espacial (representando quizás el segmento de conexión)
    geom GEOMETRY(LineString, 4326)
);
-- Índice espacial
CREATE INDEX idx_distancias_geom ON distancias USING GIST (geom);


-- -----------------------------------------------------------
-- 4. Tablas de Metadatos Espaciales (System Tables)
-- -----------------------------------------------------------

-- Nota: Estas tablas son más bien de catálogo. Usaremos la estructura que proporcionaste.

-- Tabla: geometry_columns (Catálogo de geometrías)
CREATE TABLE geometry_columns (
    f_table_catalog VARCHAR(256) NOT NULL,
    f_table_schema VARCHAR(256) NOT NULL,
    f_geometry_column VARCHAR(256) NOT NULL,
    coord_dimension INTEGER,
    srid INTEGER,
    dtype VARCHAR(30)
);

-- Tabla: geography_columns (Catálogo de geografía)
CREATE TABLE geography_columns (
    f_table_catalog VARCHAR(256) NOT NULL,
    f_table_schema VARCHAR(256) NOT NULL,
    f_geometry_column VARCHAR(256) NOT NULL,
    coord_dimension INTEGER,
    srid INTEGER,
    dtype VARCHAR(256)
);
