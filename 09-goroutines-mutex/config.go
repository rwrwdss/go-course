package main

// Config — настройки приложения (в жизни читаются из файла / env / vault — это дорого).
type Config struct {
	DSN      string
	LogLevel string
}

// ConfigProvider лениво загружает конфиг при первом обращении.
// Требование: load должен вызваться РОВНО ОДИН раз, даже если Get одновременно вызвали 100 горутин.
//
// 🐞 ЗАДАНИЕ 5: сейчас load вызывается много раз и есть гонка. Используй sync.Once.
type ConfigProvider struct {
	load func() Config
	cfg  Config
}

func NewConfigProvider(load func() Config) *ConfigProvider {
	return &ConfigProvider{load: load}
}

func (p *ConfigProvider) Get() Config {
	if p.cfg.DSN == "" {
		p.cfg = p.load()
	}
	return p.cfg
}
