package main

import (
	"bufio"
	"encoding/base64"
	"log"
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
		fmt.Printf("@G{\u2713 etcd endpoints} data in @C{%s} will be backed up\n", s)
	}

	auth, err = endpoint.BooleanValueDefault("auth", false)
	if err != nil {
		fmt.Printf("@R{\u2717 authentication %s}\n", err)
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
			fmt.Printf("@G{\u2713 username} @C{%s}\n", plugin.Redact(s))
		}

		s, err = endpoint.StringValueDefault("pass", "")
		if err != nil {
			fmt.Printf("@R{\u2717 pass  %s}\n", err)
			fail = true
		} else if s == "" {
			fmt.Printf("@R{\u2717 pass} auth was enabled but password was not provided\n")
			fail = true
		} else {
			fmt.Printf("@G{\u2713 password} @C{%s}\n", plugin.Redact(s))
		}
	}

	if fail {
		return fmt.Errorf("etcd: invalid configuration")
	}

	return nil
}

// Backup ETCD data
func (p EtcdPlugin) Backup(endpoint plugin.ShieldEndpoint) error {
	file, err := os.Open("keyval.txt")

	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var txtlines []string

	for scanner.Scan() {
		txtlines = append(txtlines, scanner.Text())
	}

	file.Close()
	i := 0
	encodedK := ""
	encodedV := ""
	for _, eachline := range txtlines {
		data := []byte(eachline)
		if i%2 == 0 {
			encodedK = base64.StdEncoding.EncodeToString(data)
		} else {
			encodedV = base64.StdEncoding.EncodeToString(data)
			newLine := fmt.Sprintf("%s : %s", encodedK, encodedV)
			cmd := fmt.Sprintf("echo %s", newLine)
			plugin.Exec(cmd, plugin.STDOUT)
			encodedK = ""
			encodedV = ""
		}
		i++
	}
	return nil
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
