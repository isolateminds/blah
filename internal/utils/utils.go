package utils

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/dannyvidal/blah/internal/color"
	"github.com/ttacon/chalk"
	"golang.org/x/term"
)

var letters = func() []rune {
	alphabet := []rune{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	for i := 'A'; i <= 'Z'; i++ {
		alphabet = append(alphabet, i)
	}
	for i := 'a'; i <= 'z'; i++ {
		alphabet = append(alphabet, i)
	}
	return alphabet
}()

//Checks to see if check string is alphanumeric a-zA-Z0-9
func IsAlphaNumeric(check string) bool {
	fn := regexp.MustCompile(`^[A-Za-z0-9]+$`).MatchString
	return fn(check)
}

//Gets user input from stdin
func GetInput(reader *bufio.Reader, output string, inputVar *string, hide bool, reason string) error {
	color.PrintForInput(output)
	if hide {
		bytepw, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return err
		}
		pass := string(bytepw)
		*inputVar = pass
		fmt.Println()
	} else {
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		input = strings.TrimSuffix(input, "\n")
		if !IsAlphaNumeric(input) {
			return fmt.Errorf("%s should be alpha numeric", reason)
		}

		*inputVar = input
	}
	return nil
}
func GenerateRandomString(length int) string {
	rand.Seed(time.Now().Unix() + rand.Int63())
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

//Iterates through functions until an error is not nil then returns error
func UntilError(fns ...func() error) error {
	for i := range fns {
		err := fns[i]()
		if err != nil {
			return err
		}
	}
	return nil
}

//Makes directory returns absolute path fatally exists if an error is encountered
func MkdirAbs(path string) string {
	err := os.Mkdir(path, os.ModePerm)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return GetAbsChild(path)
		}
		fmt.Println(chalk.Red.Color(err.Error()))
	}
	return GetAbsChild(path)
}

//Makes directory
func Mkdir(path string) {
	err := os.Mkdir(path, os.ModePerm)
	if err != nil {
		fmt.Println(chalk.Red.Color(err.Error()))
	}
}

//Returns absolute path fatally exists if an error is encountered
func GetAbsChild(path string) string {
	p, err := filepath.Abs(path)
	if err != nil {
		color.PrintFatal(err)
	}
	return p
}

//Writes a file fatally exists if an error is encountered
func WriteFile(b []byte, pathSegments ...string) {
	err := ioutil.WriteFile(path.Join(pathSegments...), b, 0666)
	if err != nil {
		color.PrintFatal(err)
	}
}

//Writes a file returns absolute path of the file fatally exists if an error is encountered
func WriteFileAbs(b []byte, pathSegments ...string) string {
	p := path.Join(pathSegments...)

	err := ioutil.WriteFile(p, b, 0666)
	if err != nil {
		color.PrintFatal(err)
	}
	return GetAbsChild(p)
}

//Prefixes project name to a container name or network name etc. EG. myproject_nginx
func PrefixProjectName(projectName string, suffix string) string {
	return fmt.Sprintf("aptcms_%s_%s", strings.ToLower(projectName), suffix)
}

//Checks if file exists
func FileExists(file string) bool {
	_, err := os.Stat(file)
	if err != nil {
		return !errors.Is(err, os.ErrNotExist)
	}
	return true
}

//Appends lines to a file if it exists
func AppendFileIfExists(fileName string, lines ...string) {
	if FileExists(fileName) {
		AppendFile(fileName, lines...)
	}
}

//Creates and appends to a file if it does not exist
func AppendFileIfNotExists(fileName string, lines ...string) {
	if !FileExists(fileName) {
		AppendFile(fileName, lines...)
	}
}

//Appends lines to a file fatally exists if an error is encountered
func AppendFile(fileName string, lines ...string) {

	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		color.PrintFatal(err)
	}
	defer f.Close()
	for i := range lines {
		_, err = f.WriteString(fmt.Sprintf("%s\n", lines[i]))
		if err != nil {
			color.PrintFatal(err)
		}
	}
}

//changes directory fatally exists if an error is encountered
func Chdir(path string) {
	if err := os.Chdir(path); err != nil {
		color.PrintFatal(err)
	}
}

//Starts a channel listening for SIGTERM Ctrl+C and invokes the callback
func HandleSIGTERM(cb func()) {
	//cleanup func upon Ctrl+C SIGINT or SIGTERM
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cb()
		os.Exit(1)
	}()
}
