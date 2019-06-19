package main

import (
	"os"

	fmt "github.com/jhunt/go-ansi"

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
		Example: `
{
  "endpoints" : "https://192.168.42.45:2379,https://192.168.42.46:2379,https://192.168.42.47:2379"   # REQUIRED

  "auth"  : false				# is role based or cert based auth enabled on the etcd cluster
  "user"  : "admin",            # username for role based authentication
  "pass"  : "p@ssw0rd"          # password for role based authentication
}
`,
		Defaults: `
{
}
`,
		Fields: []plugin.Field{
			plugin.Field{
				Mode:     "target",
				Name:     "endpoints",
				Type:     "string",
				Title:    "Endpoints",
				Help:     "IP addresses of the etcd nodes in the cluster",
				Example:  "https://192.168.42.45:2379,https://192.168.42.46:2379,https://192.168.42.47:2379",
				Required: true,
			},
			plugin.Field{
				Mode:    "target",
				Name:    "auth",
				Type:    "bool",
				Title:   "Authentication",
				Help:    "Is role based or cert based authentication enabled on the etcd cluster",
				Example: "false",
			},
			plugin.Field{
				Mode:    "target",
				Name:    "user",
				Type:    "string",
				Help:    "Username for role based authentication",
				Title:   "Username",
				Example: "admin",
			},
			plugin.Field{
				Mode:    "target",
				Name:    "pass",
				Type:    "string",
				Help:    "Password for role based authentication",
				Title:   "Password",
				Example: "p@ssw0rd",
			},
		},
	}
	fmt.Fprintf(os.Stderr, "etcd plugin starting up...\n")
	plugin.Run(p)
}

type EtcdPlugin plugin.PluginInfo

type EtcdConfig struct {
	etcdEndpoints  string
	Authentication bool
	Username       string
	Password       string
}

func (p EtcdPlugin) Meta() plugin.PluginInfo {
	return plugin.PluginInfo(p)
}

func getEtcdConfig(endpoint plugin.ShieldEndpoint) (*EtcdConfig, error) {
	etcdEndpoint, err := endpoint.StringValue("endpoints")
	if err != nil {
		return nil, err
	}

	auth, err := endpoint.BooleanValueDefault("auth", false)
	if err != nil {
		return nil, err
	}

	user, err := endpoint.StringValueDefault("user", "")
	if err != nil {
		return nil, err
	}

	pass, err := endpoint.StringValueDefault("pass", "")
	if err != nil {
		return nil, err
	}

	return &EtcdConfig{
		etcdEndpoints:  etcdEndpoint,
		Authentication: auth,
		Username:       user,
		Password:       pass,
	}, nil
}

func (p EtcdPlugin) Validate(endpoint plugin.ShieldEndpoint) error {
	var (
		s    string
		auth bool
		err  error
		fail bool
	)
	s, err = endpoint.StringValue("endpoints")
	if err != nil {
		fmt.Printf("@R{\u2717 endpoints  %s}\n", err)
		fail = true
	} else {
		fmt.Printf("@G{\u2713 endpoints}  files in @C{%s} will be backed up\n", s)
	}

	auth, err = endpoint.BooleanValueDefault("auth", false)
	if err != nil {
		fmt.Printf("@R{\u2717 authentication  %s}\n", err)
		fail = true
	} else if auth {
		fmt.Printf("@G{\u2713 auth} authentication is enabled\n")
	} else {
		fmt.Printf("@G{\u2713 auth} authentication is disabled\n")
	}

	if auth {
		s, err = endpoint.StringValueDefault("user", "")
		if err != nil {
			fmt.Printf("@R{\u2717 user  %s}\n", err)
			fail = true
		} else if s == "" {
			fmt.Printf("@G{\u2717 user} auth was enabled but username was not provided\n")
			fail = true
		} else {
			fmt.Printf("@G{\u2713 user} auth will be done with username @C{%s}\n", s)
		}

		s, err = endpoint.StringValueDefault("pass", "")
		if err != nil {
			fmt.Printf("@R{\u2717 pass  %s}\n", err)
			fail = true
		} else if s == "" {
			fmt.Printf("@R{\u2717 pass} auth was enabled but password was not provided\n")
			fail = true
		} else {
			fmt.Printf("@G{\u2713 user} auth will be done with password @C{%s}\n", s)
		}
	}

	if fail {
		return fmt.Errorf("fs: invalid configuration")
	}

	return nil
}

func (p EtcdPlugin) Backup(endpoint plugin.ShieldEndpoint) error {
	return plugin.UNIMPLEMENTED
}

func (p EtcdPlugin) Restore(endpoint plugin.ShieldEndpoint) error {
	return plugin.UNIMPLEMENTED
}

func (p EtcdPlugin) Store(endpoint plugin.ShieldEndpoint) (string, int64, error) {
	return "", 0, plugin.UNIMPLEMENTED
}

func (p EtcdPlugin) Retrieve(endpoint plugin.ShieldEndpoint, file string) error {
	return plugin.UNIMPLEMENTED
}

func (p EtcdPlugin) Purge(endpoint plugin.ShieldEndpoint, key string) error {
	return plugin.UNIMPLEMENTED
}
