package rmm

import (
	"rahnit-rmm/pki"
	"rahnit-rmm/util"
	"sync"
)

type ConfigManager struct {
	configs map[string]util.ObservableMap[string, pki.SignedArtifact[HostConfig]]
	mutex   sync.Mutex
}

func (cm *ConfigManager) GetConfigMap(configType string) util.ObservableMap[string, pki.SignedArtifact[HostConfig]] {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

}
