package config

type AppConfig struct {
	KafkaConf  `ini:"kafka"`
	EtcdConfig `ini:"etcd"`
}

type KafkaConf struct {
	Address     string `ini:"address"`
	Topic       string `ini:"topic"`
	ChanMaxSize int    `ini:"chan_max_size"`
}
type EtcdConfig struct {
	Address       string `ini:"address"`
	Timeout       int    `ini:"timeout"`
	CollectLogKey string `ini:"collect_log_key"`
}
type TailflogConf struct {
	Path string `ini:"path"`
}
