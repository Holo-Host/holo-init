package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"reflect"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/badoux/checkmail"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

type Contact struct {
	Id       int64  `json:"id"`
	Response string `json:"response"`
}
type Update struct {
	ContactId int64  `json:"contact_id"`
	PublicKey string `json:"public_key"`
}

type Dns struct {
	Pubkey string `json:"pubkey"`
}

type ProxyService struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

type ProxyRoute struct {
	Name      string   `json:"name"`
	Protocols []string `json:"protocols"`
	Hosts     []string `json:"hosts"`
	Service   string   `json:"service"`
}

type agents struct {
	Id             string `toml:"id"`
	Name           string `toml:"name"`
	Keystore_file  string `toml:"keystore_file"`
	Public_address string `toml:"public_address"`
}

type bridges struct {
	CallerId string `toml:"caller_id"`
	CalleeId string `toml:"callee_id"`
	Handle   string `toml:"handle"`
}

type dnas struct {
	File string `toml:"file"`
	Id   string `toml:"id"`
}

type driver struct {
	Port int    `toml:"port"`
	Type string `toml:"type"`
}

type storage struct {
	Path string `toml:"path"`
	Type string `toml:"type"`
}

type instances struct {
	Agent   string   `toml:"agent"`
	Dna     string   `toml:"dna"`
	Id      string   `toml:"id"`
	Storage *storage `toml:"storage"`
}
type intefaces_instances struct {
	Id string `toml:"id"`
}
type interfaces struct {
	Id        string                `toml:"id"`
	Admin     bool                  `toml:"admin"`
	Driver    *driver               `toml:"driver"`
	Instances []intefaces_instances `toml:"instances"`
}

type network struct {
	Bootstrap_nodes      []string `toml:"bootstrap_nodes"`
	N3h_persistence_path string   `toml:"n3h_persistence_path"`
}

type ui_bundles struct {
	Id       string `toml:"id"`
	Hash     string `toml:"hash"`
	Root_dir string `toml:"root_dir"`
}

type ui_interfaces struct {
	Bundle        string `toml:"bundle"`
	Dna_interface string `toml:"dna_interface"`
	Id            string `toml:"id"`
	Port          int    `toml:"port"`
}

type Config struct {
	PD             string          `toml:"persistence_dir"`
	Agents         []agents        `toml:"agents"`
	Bridges        []bridges       `toml:"bridges"`
	Dnas           []dnas          `toml:"dnas"`
	Instances      []instances     `toml:"instances"`
	Interfaces     []interfaces    `toml:"interfaces"`
	SigningService string          `toml:"signing_service_uri"`
	Network        network         `toml:"network"`
	UiBundles      []ui_bundles    `toml:"ui_bundles"`
	UiInterfaces   []ui_interfaces `toml:"ui_interfaces"`
}

func init() {
	rootCmd.AddCommand(initCmd)
}
func LsCmd() string {
	lscmd := exec.Command("sudo", "-u", "holochain", "ls", "/home/holochain/.config/holochain/keys")
	var stdout, stderr bytes.Buffer
	lscmd.Stdout = &stdout
	lscmd.Stderr = &stderr
	lserr := lscmd.Run()
	if lserr != nil {
		fmt.Println(lserr)
	}
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	fmt.Printf("current keys:\n%s\n", outStr)
	if errStr != "" {
		fmt.Println(errStr)
	}
	return outStr
}
func ZtListNet() string {
	lscmd := exec.Command("sudo", "zerotier-cli", "listnetworks", "-j")
	var stdout, stderr bytes.Buffer
	lscmd.Stdout = &stdout
	lscmd.Stderr = &stderr
	lserr := lscmd.Run()
	if lserr != nil {
		fmt.Println(lserr)
	}
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if errStr != "" {
		fmt.Println(errStr)
	}
	return outStr
}
func HcKeyGen() {
	outStr := LsCmd()
	matched, merr := regexp.MatchString(`Hc.*`, outStr)
	if merr != nil {
		fmt.Println(merr)
	}
	if matched != true {
		cmd := exec.Command("sudo", "-u", "holochain", "hc", "keygen", "-n")
		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err := cmd.Run()
		if err != nil {
			fmt.Println(errb.String())
		}
		fmt.Println("out:", outb.String(), "err:", errb.String())
	} else {
		fmt.Println("Your already have existing keys")
	}
	return
}
func ZtAuth() {
	ztcmd := exec.Command("sudo", "zerotier-cli", "info", "-j")
	var ztstdout, ztstderr bytes.Buffer
	ztcmd.Stdout = &ztstdout
	ztcmd.Stderr = &ztstderr
	zterr := ztcmd.Run()
	if zterr != nil {
		fmt.Println(zterr)
	}

	ztresult := ztstdout.Bytes()
	ztdata := make(map[string]interface{})
	jerr := json.Unmarshal(ztresult, &ztdata)
	if jerr != nil {
		fmt.Println(jerr)
	}
	fmt.Println(ztdata["address"].(string))
	message := map[string]interface{}{
		"member_id": ztdata["address"],
	}

	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		log.Fatalln(err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://proxy.holohost.net/zato/holo-zt-auth", bytes.NewBuffer(bytesRepresentation))
	req.Header.Add("Host", "proxy.holohost.net")
	req.Header.Add("Holo-Init", "wbfGXvzmLk83bUmR")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	var thisresult map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&thisresult)
	log.Println(thisresult)

}
func EmailAddress() {

	validate := func(input string) error {
		checkerr := checkmail.ValidateFormat(input)
		if checkerr != nil {
			return errors.New("invalid email")
		}
		return nil
	}

	var username string
	prompt := promptui.Prompt{
		Label:    "Email Address",
		Validate: validate,
		Default:  username,
	}

	result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	fmt.Printf("Your email address is %q\n", result)
	message := map[string]interface{}{
		"email": result,
	}

	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		log.Fatalln(err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://proxy.holohost.net/zato/holo-init-email", bytes.NewBuffer(bytesRepresentation))
	req.Header.Add("Host", "proxy.holohost.net")
	req.Header.Add("Holo-Init", "wbfGXvzmLk83bUmR")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	var thisresult Contact
	json.NewDecoder(resp.Body).Decode(&thisresult)
	// log.Println(thisresult.Id)
	// log.Println(thisresult.Response)
	if thisresult.Response != "" {
		log.Println(thisresult.Response)
	}
	if thisresult.Id != 0 {
		pk := LsCmd()
		//fmt.Println(reflect.TypeOf(thisresult.Id))
		regmessage := &Update{
			ContactId: thisresult.Id,
			PublicKey: pk,
		}

		regbytesRepresentation, brerr := json.Marshal(regmessage)
		if brerr != nil {
			fmt.Println(brerr)
		}

		reg, rerr := http.NewRequest("POST", "http://proxy.holohost.net/zato/holo-init-update", bytes.NewBuffer(regbytesRepresentation))
		reg.Header.Add("Host", "proxy.holohost.net")
		reg.Header.Add("Holo-Init", "wbfGXvzmLk83bUmR")
		regresp, rerr := client.Do(reg)
		if rerr != nil {
			fmt.Println(rerr)
		}
		var regresult map[string]interface{}
		json.NewDecoder(regresp.Body).Decode(&regresult)
		log.Println(regresult)

	}

}
func CloudFlare() {
	//var publicKey string
	var match []string
	var ipaddress string
	publicKey := LsCmd()
	//fmt.Println(LsCmd())
	fmt.Println(ZtListNet())
	ztcmd := ZtListNet()
	var ztdata []interface{}
	//var ztipv4 string
	jerr := json.Unmarshal([]byte(ztcmd), &ztdata)
	if jerr != nil {
		fmt.Println(jerr)
	}
	fmt.Println(reflect.TypeOf(ztcmd))
	//pattern := regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)
	pattern := regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)

	for _, i := range ztdata {
		if rec, ok := i.(map[string]interface{}); ok {
			for key, val := range rec {
				if key == "nwid" && val == "93afae5963c547f1" {
					log.Printf(" [========>] %s = %s", key, val)
					for _, item := range rec["assignedAddresses"].([]interface{}) {
						if str, ok := item.(string); ok {
							//fmt.Println(pattern.MatchString(str))
							if pattern.MatchString(str) == true {
								//ztipv4 = str
								match = pattern.FindStringSubmatch(str)
								for _, element := range match {
									element = strings.Trim(element, "[")
									element = strings.Trim(element, "]")
									ipaddress = element
									//fmt.Println(ipaddress)
								}
							}
						} else {
							//fmt.Println(item)
						}
						//fmt.Println(reflect.TypeOf(item).Kind())

					}
				}
			}
		} else {
			fmt.Printf("Error with this data: %v\n", i)
		}
	}
	fmt.Println(ipaddress)
	publicKey = strings.TrimSuffix(publicKey, "\n")
	dnsclient := &http.Client{}
	dnsmessage := &Dns{
		Pubkey: publicKey,
	}

	dnsbytesRepresentation, brerr := json.Marshal(dnsmessage)
	if brerr != nil {
		fmt.Println(brerr)
	}

	dnsreg, dnsrerr := http.NewRequest("POST", "http://proxy.holohost.net/zato/holo-init-cloudflare-dns-create", bytes.NewBuffer(dnsbytesRepresentation))
	dnsreg.Header.Add("Host", "proxy.holohost.net")
	dnsreg.Header.Add("Content-Type", "application/json")
	dnsreg.Header.Add("Holo-Init", "wbfGXvzmLk83bUmR")
	regresp, dnsrerr := dnsclient.Do(dnsreg)
	if dnsrerr != nil {
		fmt.Println(dnsrerr)
	}
	var regresult map[string]interface{}
	json.NewDecoder(regresp.Body).Decode(&regresult)
	log.Println(regresult)

	proxyclient := &http.Client{}
	proxymessage := &ProxyService{
		Name:     publicKey + ".holohost.net",
		Protocol: "http",
		Host:     ipaddress,
		Port:     48080,
	}

	proxybytesRepresentation, proxyerr := json.Marshal(proxymessage)
	if proxyerr != nil {
		fmt.Println(proxyerr)
	}
	log.Println("TRYING PROXY SERVICE CREATE")
	proxyreg, proxyerr := http.NewRequest("POST", "http://proxy.holohost.net/zato/holo-init-proxy-service-create", bytes.NewBuffer(proxybytesRepresentation))
	proxyreg.Header.Add("Host", "proxy.holohost.net")
	proxyreg.Header.Add("Content-Type", "application/json")
	proxyreg.Header.Add("Holo-Init", "wbfGXvzmLk83bUmR")
	proxyregresp, proxyerr := proxyclient.Do(proxyreg)
	if proxyerr != nil {
		fmt.Println(proxyerr)
	}
	var proxyresult map[string]interface{}
	json.NewDecoder(proxyregresp.Body).Decode(&proxyresult)
	log.Println(proxyresult["id"])
	var serviceId string
	if sidstr, ok := proxyresult["id"].(string); ok {
		serviceId = sidstr
	} else {

	}
	proxyrclient := &http.Client{}
	var downcasepubkey string
	downcasepubkey = strings.ToLower(publicKey)
	proxyrmessage := &ProxyRoute{
		Name:      publicKey + ".holohost.net",
		Protocols: []string{"http", "https"},
		Hosts:     []string{"*." + downcasepubkey + ".holohost.net"},
		Service:   serviceId,
	}

	proxyrbytesRepresentation, proxyrerr := json.Marshal(proxyrmessage)
	if proxyrerr != nil {
		fmt.Println(proxyrerr)
	}
	log.Println("TRYING PROXY ROUTE CREATE")
	proxyrreg, proxyrerr := http.NewRequest("POST", "http://proxy.holohost.net/zato/holo-proxy-route-create", bytes.NewBuffer(proxyrbytesRepresentation))
	proxyrreg.Header.Add("Host", "proxy.holohost.net")
	proxyrreg.Header.Add("Content-Type", "application/json")
	proxyrreg.Header.Add("Holo-Init", "wbfGXvzmLk83bUmR")
	proxyrregresp, proxyrerr := proxyrclient.Do(proxyrreg)
	if proxyrerr != nil {
		fmt.Println(proxyrerr)
	}
	var proxyrresult map[string]interface{}
	json.NewDecoder(proxyrregresp.Body).Decode(&proxyrresult)
	log.Println(proxyrresult)
	//fmt.Println(ztipv4)
	//fmt.Println(ipaddress)
	//fmt.Println(publicKey)

}
func ConfigRewrite() {
	var config Config
	if _, err := toml.DecodeFile("/var/lib/holochain/conductor-config.toml", &config); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("persistence_dir: %s \n agents: %s \n bridges: %s \n dnas: %s \n instances: %s \n interfaces: %s \n network: %s \n signing_service: %s \n ui_interfaces: %s \n ui_bundles: %s \n",
		config.PD, config.Agents, config.Bridges, config.Dnas, config.Instances, config.Interfaces, config.Network, config.SigningService, config.UiInterfaces, config.UiBundles)
	fmt.Println(LsCmd())
	buf := new(bytes.Buffer)
	var signing_service string
	signing_service = "http://localhost:8888"
	publicKey := LsCmd()
	publicKey = strings.TrimSuffix(publicKey, "\n")
	initconf := Config{
		PD:             config.PD,
		SigningService: signing_service,
		Agents: []agents{
			{
				Id:             "host-agent",
				Name:           "Envoy Host",
				Keystore_file:  "/home/holochain/.config/holochain/keys/" + publicKey,
				Public_address: publicKey,
			},
		},
		Bridges: config.Bridges,
		Network: network{
			//Bootstrap_nodes: [],
			N3h_persistence_path: "/home/holochain/.n3h",
		},
		Dnas: []dnas{
			{
				File: "/run/current-system/sw/bin/envoy/src/config/.envoy-deps/holo-hosting.dna.json",
				Id:   "holo-hosting-app",
			},
			{
				File: "/run/current-system/sw/bin/envoy/src/config/.envoy-deps/happ-store.dna.json",
				Id:   "happ-store",
			},
			{
				File: "/run/current-system/sw/bin/envoy/src/config/.envoy-deps/holofuel.dna.json",
				Id:   "holofuel",
			},
		},
		Instances: []instances{
			{
				Agent: "host-agent",
				Dna:   "holo-hosting-app",
				Id:    "holo-hosting-app",
				Storage: &storage{
					Path: "/var/lib/holochain/storage/holo-hosting-app",
					Type: "file",
				},
			},
			{
				Agent: "host-agent",
				Dna:   "happ-store",
				Id:    "happ-store",
				Storage: &storage{
					Path: "/var/lib/holochain/storage/happ-store",
					Type: "file",
				},
			},
			{
				Agent: "host-agent",
				Dna:   "holofuel",
				Id:    "holofuel",
				Storage: &storage{
					Path: "/var/lib/holochain/storage/holofuel",
					Type: "file",
				},
			},
		},
		Interfaces: []interfaces{
			{
				Id:    "master-interface",
				Admin: true,
				Instances: []intefaces_instances{
					{Id: "holo-hosting-app"},
					{Id: "happ-store"},
				},
				Driver: &driver{
					Port: 1111,
					Type: "websocket",
				},
			},
			{
				Id: "public-interface",
				Driver: &driver{
					Port: 2222,
					Type: "websocket",
				},
			},
			{
				Id: "internal-interface",
				Driver: &driver{
					Port: 3333,
					Type: "websocket",
				},
			},
		},
		UiBundles: []ui_bundles{
			{
				Hash:     "Qm000",
				Id:       "hha-ui",
				Root_dir: "/run/current-system/sw/bin/envoy/src/config/.envoy-deps/holo-hosting-ui",
			},
			{
				Hash:     "Qm001",
				Id:       "happ-store-ui",
				Root_dir: "/run/current-system/sw/bin/envoy/src/config/.envoy-deps/happ-store-ui",
			},
		},
		UiInterfaces: []ui_interfaces{
			{
				Bundle:        "hha-ui",
				Dna_interface: "master-interface",
				Id:            "hha-ui-interface",
				Port:          8800,
			},
			{
				Bundle:        "happ-store-ui",
				Dna_interface: "master-interface",
				Id:            "happ-store-ui-interface",
				Port:          8880,
			},
		},
	}
	if encerr := toml.NewEncoder(buf).Encode(initconf); encerr != nil {
		log.Println(encerr)
	}
	fmt.Println(buf.String())
	message := []byte(buf.String())
	werr := ioutil.WriteFile("/var/lib/holochain/conductor-config.toml", message, 0775)
	if werr != nil {
		log.Println(werr)
	}
	cmd := exec.Command("sudo", "-u", "root", "systemctl", "restart", "holochain.service")
	// cmd.Stdout = &outb
	// cmd.Stderr = &errb
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cerr := cmd.Run()
	if cerr != nil {
		fmt.Println(cerr)
	}
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	fmt.Printf("\n%s\n", outStr)
	fmt.Println("Your Holoport is now initialized")
	if errStr != "" {
		fmt.Println(errStr)
	}

}

//func PassWord(){
//	validatepw := func(input string) error {
//		if len(input) < 6 {
//			return errors.New("Password must have more than 6 characters")
//		}
//		return nil
//	}
//
//	prompt := promptui.Prompt{
//		Label:    "Password",
//		Validate: validatepw,
//		Mask:     '*',
//	}
//
//	result1, err := prompt.Run()
//
//	if err != nil {
//		fmt.Printf("Prompt failed %v\n", err)
//		return
//	}
//
//	fmt.Printf("Your password is %q\n", result1)
//}
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "The initialization command",
	Long:  `Get your Holoport up and running`,
	Run: func(cmd *cobra.Command, args []string) {
		HcKeyGen()
		ZtAuth()
		EmailAddress()
		CloudFlare()
		ConfigRewrite()
		//PassWord()
		//fmt.Println("Test123")
	},
}
