package cmd

import (
  "fmt"
  "bytes"
  //"os/exec"
  "net/http"
  "log"
  //"io/ioutil"
  "encoding/json"
  //"strconv"
  "errors" 
  "github.com/spf13/cobra"
  "github.com/manifoldco/promptui"
)
//type pepper struct {
//	Name     string
//	HeatUnit int
//	Peppers  int
//}
func init() {
  rootCmd.AddCommand(initCmd)
}

func MakeRequest() {

	message := map[string]interface{}{
		"hello": "world",
		"life":  42,
		"embedded": map[string]string{
			"yes": "of course!",
		},
	}

	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post("https://httpbin.org/post", "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)
	}

	var result map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&result)

	log.Println(result)
	log.Println(result["data"])
}
func EmailAddress(){
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

	resp, err := http.Post("https://httpbin.org/post", "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)
	}

	var thisresult map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&thisresult)

	log.Println(thisresult)


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
    //MakeRequest()
    EmailAddress()
    //PassWord()
    fmt.Println("Test123")
  },
}
