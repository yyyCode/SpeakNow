package config

// BuiltinDefault 在 go:embed 未打入配置时的兜底（与 configs/config.release.yaml 一致）。
func BuiltinDefault() []byte {
	return []byte(`server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"
  read_header_timeout: 10s
  read_timeout: 0s
  write_timeout: 0s
  idle_timeout: 120s
  shutdown_timeout: 30s
cache:
  ttl: 24h
ratelimit:
  global_qps: 100
  global_burst: 200
  user_qps: 5
  user_burst: 10
asr:
  primary: "vosk"
  fallback: []
  timeout: 60s
  max_audio_size: 10485760
  max_duration: 300s
providers:
  mock:
    enabled: false
  aliyun:
    enabled: false
  tencent:
    enabled: false
  xunfei:
    enabled: false
  vosk:
    enabled: true
    model_path: "embedded"
    sample_rate: 16000
    cost_per_second: 0
log:
  level: "info"
  output: "stdout"
`)
}
