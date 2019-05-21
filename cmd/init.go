package cmd

import (
	"bytes"
	"fmt"

	"log"
	"net/http"
	"regexp"

	"encoding/json"
	"errors"
	"os/exec"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

func HcKeyGen() {
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
	matched, merr := regexp.MatchString(`Hc.*`, outStr)
	if merr != nil {
		fmt.Println(merr)
	}
	if matched != true {
		cmd := exec.Command("sudo", "-u", "holochain", "hc", "keygen", "-n")
		//cmd.SysProcAttr = &syscall.SysProcAttr{}
		//cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(401), Gid: uint32(501), NoSetGroups: false}
		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err := cmd.Run()
		if err != nil {
			//log.Fatal(errb.String())
		}
		fmt.Println("out:", outb.String(), "err:", errb.String())
	} else {
		fmt.Println("Your already have existing keys")
	}
	return
}
func ZtAuth() {}
func EmailAddress() {
	validate := func(input string) error {
		if len(input) < 3 {
			return errors.New("email address must have more than 3 characters")
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

	// resp, err := http.Post("https://httpbin.org/post", "application/json", bytes.NewBuffer(bytesRepresentation))
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://proxy.holohost.net/zato/holo-init-email", bytes.NewBuffer(bytesRepresentation))
	req.Header.Add("Host", "proxy.holohost.net")
	req.Header.Add("Holo-Init", "wbfGXvzmLk83bUmR")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	// bodyText, err := ioutil.ReadAll(resp.Body)
	// s := string(bodyText)
	var thisresult map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&thisresult)
	if _, ok := thisresult["response"]; ok {
		log.Println(thisresult["response"])
	}
	if _, ok := thisresult["id"]; ok {
		//i := strconv.Atoi(thisresult["id"])
		log.Println(thisresult["id"])
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
