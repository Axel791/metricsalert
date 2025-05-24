package shared

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"
)

// LoadEnvFromFile читает пары "VAR":"value" из JSON-файла и
// выставляет их в окружение. Если файла нет — молча возвращает.
func LoadEnvFromFile(logger *logrus.Logger, path string) {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		logger.Warnf("cannot read %s: %v", path, err)
		return
	}
	var kv map[string]string
	if err = json.Unmarshal(data, &kv); err != nil {
		logger.Warnf("cannot parse %s: %v", path, err)
		return
	}
	for k, v := range kv {
		_ = os.Setenv(k, v)
	}
	logger.Infof("environment loaded from %s", path)
}
