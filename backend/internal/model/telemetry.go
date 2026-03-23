package model

import (
	"errors"
	"strings"
	"time"
)

type PacoteTelemetria struct {
	IDDispositivo  string `json:"id_dispositivo"`
	Instante       string `json:"instante"`
	TipoSensor     string `json:"tipo_sensor"`
	NaturezaLeitura string `json:"natureza_leitura"`
	ValorColetado  string `json:"valor_coletado"`
}

func (p PacoteTelemetria) Validar() error {
	if strings.TrimSpace(p.IDDispositivo) == "" {
		return errors.New("id_dispositivo eh obrigatorio")
	}

	if strings.TrimSpace(p.Instante) == "" {
		return errors.New("instante eh obrigatorio")
	}

	if _, err := time.Parse(time.RFC3339, p.Instante); err != nil {
		return errors.New("instante precisa estar em RFC3339")
	}

	if strings.TrimSpace(p.TipoSensor) == "" {
		return errors.New("tipo_sensor eh obrigatorio")
	}

	switch strings.ToLower(strings.TrimSpace(p.NaturezaLeitura)) {
	case "analogico", "discreto":
	default:
		return errors.New("natureza_leitura precisa ser analogico ou discreto")
	}

	if strings.TrimSpace(p.ValorColetado) == "" {
		return errors.New("valor_coletado eh obrigatorio")
	}

	return nil
}
