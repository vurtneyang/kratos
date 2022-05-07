package txmq

const (
	stateRunning int32 = 1
)

type Config struct {
	ServerAddress string
	AccessKey     string
	SecretKey     string
	NameSpace     string
}
