import http from "k6/http";
import { check, sleep } from "k6";

const maxVUs = Number(__ENV.MAX_VUS || 30);
const rampSeconds = Number(__ENV.RAMP_SECONDS || 30);
const holdSeconds = Number(__ENV.HOLD_SECONDS || 60);
const downSeconds = Number(__ENV.RAMP_DOWN_SECONDS || 30);
const sleepSeconds = Number(__ENV.SLEEP_SECONDS || 0.1);

export const options = {
  stages: [
    { duration: `${rampSeconds}s`, target: maxVUs },
    { duration: `${holdSeconds}s`, target: maxVUs },
    { duration: `${downSeconds}s`, target: 0 },
  ],
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<500"],
  },
};

const baseUrl = __ENV.BASE_URL || "http://localhost:8080";
const tiposSensor = ["temperatura", "umidade", "presenca", "vibracao", "luminosidade", "nivel_tanque"];

function randomInt(max) {
  return Math.floor(Math.random() * max);
}

export default function () {
  const tipoSensor = tiposSensor[randomInt(tiposSensor.length)];
  const naturezaLeitura = tipoSensor === "presenca" ? "discreto" : "analogico";
  const valorColetado = naturezaLeitura === "discreto" ? (Math.random() > 0.5 ? "ligado" : "desligado") : `${(Math.random() * 100).toFixed(2)}`;

  const payload = JSON.stringify({
    id_dispositivo: `disp-${__VU}-${__ITER}`,
    instante: new Date().toISOString(),
    tipo_sensor: tipoSensor,
    natureza_leitura: naturezaLeitura,
    valor_coletado: valorColetado,
  });

  const params = {
    headers: { "Content-Type": "application/json" },
  };

  const res = http.post(`${baseUrl}/telemetria`, payload, params);
  check(res, {
    "status eh 202": (r) => r.status === 202,
  });

  if (sleepSeconds > 0) {
    sleep(sleepSeconds);
  }
}
