package main

import (
	"fmt"
	"os"

	"github.com/starkandwayne/shield/plugin"
)

func main() {
	p := EtcdPlugin{
		Name:    "Etcd Backup Plugin",
		Author:  "Stark and Wayne",
		Version: "0.0.1",
		Features: plugin.PluginFeatures{
			Target: "yes",
			Store:  "no",
		},
	}
	fmt.Fprintf(os.Stderr, "etcd plugin starting up...\n")
	plugin.Run(p)
}

type EtcdPlugin plugin.PluginInfo

type EtcdConnectionInfo struct {
	Host string
	Path string
}

func (p EtcdPlugin) Meta() plugin.PluginInfo {
	return plugin.PluginInfo(p)
}

func (p EtcdPlugin) Validate(endpoint plugin.ShieldEndpoint) error {
	fmt.Println("Validate function called")
	return plugin.UNIMPLEMENTED
}

func (p EtcdPlugin) Backup(endpoint plugin.ShieldEndpoint) error {
	fmt.Println("Backup function called")
	return plugin.UNIMPLEMENTED
}

func (p EtcdPlugin) Restore(endpoint plugin.ShieldEndpoint) error {
	fmt.Println("Restore function called")
	return plugin.UNIMPLEMENTED
}

func (p EtcdPlugin) Store(endpoint plugin.ShieldEndpoint) (string, int64, error) {
	fmt.Println("Store function called")
	return "", 0, plugin.UNIMPLEMENTED
}

func (p EtcdPlugin) Retrieve(endpoint plugin.ShieldEndpoint, file string) error {
	fmt.Println("Retrieve function called")
	return plugin.UNIMPLEMENTED
}

func (p EtcdPlugin) Purge(endpoint plugin.ShieldEndpoint, key string) error {
	fmt.Println("Purge function called")
	return plugin.UNIMPLEMENTED
}
