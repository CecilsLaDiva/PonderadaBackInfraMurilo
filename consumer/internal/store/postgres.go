package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RegistroTelemetria struct {
	IDDispositivo   string `json:"id_dispositivo"`
	Instante        string `json:"instante"`
	TipoSensor      string `json:"tipo_sensor"`
	NaturezaLeitura string `json:"natureza_leitura"`
	ValorColetado   string `json:"valor_coletado"`
}

type TelemetryStore interface {
	Insert(context.Context, RegistroTelemetria) error
	Close()
}

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(databaseURL string) (*PostgresStore, error) {
	var pool *pgxpool.Pool
	var err error

	for i := 0; i < 12; i++ {
		pool, err = pgxpool.New(context.Background(), databaseURL)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			pingErr := pool.Ping(ctx)
			cancel()
			if pingErr == nil {
				return &PostgresStore{pool: pool}, nil
			}
			err = pingErr
			pool.Close()
		}
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("erro ao conectar no postgres: %w", err)
}

func DecodificarTelemetria(body []byte) (RegistroTelemetria, error) {
	var record RegistroTelemetria
	if err := json.Unmarshal(body, &record); err != nil {
		return RegistroTelemetria{}, fmt.Errorf("erro ao decodificar msg da fila: %w", err)
	}
	if _, err := time.Parse(time.RFC3339, record.Instante); err != nil {
		return RegistroTelemetria{}, fmt.Errorf("instante com formato invalido: %w", err)
	}
	return record, nil
}

func (s *PostgresStore) Insert(ctx context.Context, record RegistroTelemetria) error {
	observadoEm, err := time.Parse(time.RFC3339, record.Instante)
	if err != nil {
		return fmt.Errorf("erro parseando instante antes do insert: %w", err)
	}

	query := `
		INSERT INTO leituras_telemetria (id_dispositivo, observado_em, tipo_sensor, natureza_leitura, valor_coletado)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = s.pool.Exec(
		ctx,
		query,
		record.IDDispositivo,
		observadoEm,
		record.TipoSensor,
		record.NaturezaLeitura,
		record.ValorColetado,
	)
	if err != nil {
		return fmt.Errorf("erro inserindo leitura no banco: %w", err)
	}
	return nil
}

func (s *PostgresStore) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}
