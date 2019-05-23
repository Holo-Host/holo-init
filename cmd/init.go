package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"regexp"

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
	//outStr, errStr := string(ztstdout.Bytes()), string(ztstderr.Bytes())
	//fmt.Printf("zt json:\n%s\n", outStr)
	//fmt.Println(reflect.TypeOf(outStr).Kind())
	// if errStr != "" {
	// 	fmt.Println(errStr)
	// }

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
		//PassWord()
		//fmt.Println("Test123")
	},
}
