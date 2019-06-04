package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"reflect"
	"regexp"
	"strings"

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

	dnsreg, dnsrerr := http.NewRequest("POST", "http://proxy.holohost.net/zato/holo-cloudflare-dns-create", bytes.NewBuffer(dnsbytesRepresentation))
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
	proxyreg, proxyerr := http.NewRequest("POST", "http://proxy.holohost.net/zato/holo-proxy-service-create", bytes.NewBuffer(proxybytesRepresentation))
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
		//PassWord()
		//fmt.Println("Test123")
	},
}
