CREATE TABLE IF NOT EXISTS leituras_telemetria (
    id BIGSERIAL PRIMARY KEY,
    id_dispositivo TEXT NOT NULL,
    observado_em TIMESTAMPTZ NOT NULL,
    tipo_sensor TEXT NOT NULL,
    natureza_leitura TEXT NOT NULL CHECK (natureza_leitura IN ('analogico', 'discreto')),
    valor_coletado TEXT NOT NULL,
    criado_em TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_telemetria_criado_em
    ON leituras_telemetria (criado_em DESC);

CREATE INDEX IF NOT EXISTS idx_telemetria_id_dispositivo
    ON leituras_telemetria (id_dispositivo);
