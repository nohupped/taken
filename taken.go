package main

import (
	"github.com/rjeczalik/notify"
	"log"
	"fmt"
	"path/filepath"
	"os"
	"strings"
	"bytes"
	"os/exec"
	."faimodules"
	"bufio"
	"io"
	"os/user"
)

func checkifdir(file string) bool  {
	filestat, _ := os.Stat(file)
	if filestat.IsDir(){
		return true
	}else {
		return false
	}
}

func getHomes(home string) ([]string, []string) {
	homepaths, _ := filepath.Glob(home)
	homedirs := make([]string, 0, 1)
	usernames := make([]string, 0, 1)
	for _, i := range homepaths {
		if checkifdir(i) {
			homedirs = append(homedirs, i)
			tmp := strings.Split(i, "/")[2]
			usernames = append(usernames, tmp)
		}
	}
	return homedirs, usernames

}

func ValidateDeb(filename string) bool{

	cmd := exec.Command("/usr/bin/file", filename)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(err, stderr.String())
		Error.Println(err, stderr.String())
		panic(err)
	}
	Info.Println(out.String())
	return strings.Contains(out.String(), "Debian binary package")



}
func moveDebs(src, dst string) string {
	srcsplit := strings.Split(src, "/")
	filename := srcsplit[len(srcsplit) -1]
	dstpath := dst + filename
	fi, err := os.Open(src)
	if err != nil {
		Error.Println("Error while moving: ", err)
	}
	defer fi.Close()
	r := bufio.NewReader(fi)
	fo, err := os.Create(dstpath)
	if err != nil {
		Error.Println("Error while moving: ", err)
	}
	defer fo.Close()
	w := bufio.NewWriter(fo)
	buf := make([]byte, 1024)
	for {
		// read a chunk
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}
		if _, err := w.Write(buf[:n]); err != nil {
			panic(err)
		}
	}
	if err = w.Flush(); err != nil {
		Error.Println(err)
	}
	os.Remove(src)
	Info.Println("Moved deb files.", src, "-->", dstpath)
	return dstpath

}


func CallAptlyAdd(debfile string) {
	Info.Println("Calling aptly add")
	bin := "sudo"
	args := []string{"-u", "repo", "aptly", "repo", "add", "Sysops", debfile}
	cmd := exec.Command( bin, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		Error.Println(err, out.String())
		Error.Println(err, stderr.String())
	}
	Info.Println(out.String())
	os.Remove(debfile)
	Info.Println("Removed", debfile)

}

func CallAptlyShow() {
	Info.Println("Calling aptly show")
	cmd := exec.Command( "sudo", "-u", "repo", "aptly", "repo", "show", "-with-packages", "Sysops")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		Error.Println(err, out.String())
		Error.Println(err, stderr.String())
	}
	Info.Println(out.String())
}

func CallAptlyPublish()  {
	Info.Println("Calling aptly publish")
	cmd := exec.Command( "sudo", "-u", "repo", "aptly", "--skip-signing", "publish", "update", "testing")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		Error.Println(err, stderr.String())
		Error.Println(err, out.String())
	}
	Info.Println(out.String())

}



func main() {

	repotmp := "/repo/PKGBUILD/"
	currentuser, _ := user.Current()
	homedirs, _ := getHomes("/home/*")
	logFile := StartLog("/var/log/pkgbuild.log", currentuser)
	defer logFile.Close()
	Info.Println("\n\n\t\tI don't know who you are\n\t\tI don't know what you are syncing\n\t\tIf you are syncing via rsync, I can tell you \n\t\tI don't have the condition to pick it\n\t\tBut what I do have are a very particular set of channels\n\t\tChannels that pick up debs and push it to the repo\n\t\tI will look for debs in", homedirs, ", \n\t\tI will find it, and I will add it to repo . . .\n")
	fmt.Println("\n\n\t\tI don't know who you are\n\t\tI don't know what you are syncing\n\t\tIf you are syncing via rsync, I can tell you \n\t\tI don't have the condition to pick it\n\t\tBut what I do have are a very particular set of channels\n\t\tChannels that pick up debs and push it to the repo\n\t\tI will look for debs in", homedirs, ", \n\t\tI will find it, and I will add it to repo . . .\n")
	Info.Println("Starting taken . . .\nUse scp to copy deb files.")
	fmt.Println("Starting taken . . .\nUse scp to copy deb files.")
	Info.Println("Deb files will be moved to ", repotmp, "before pushing to repo")


	c := make(chan notify.EventInfo, 20)


	for _, i := range homedirs {
		if err := notify.Watch(i, c, notify.InCloseWrite, notify.Create); err != nil {
			log.Fatal(err)
		}
	}
	/*if err := notify.Watch("/home/girishg/", c, notify.InCloseWrite, notify.All); err != nil {
		log.Fatal(err)
	}
	if err := notify.Watch("/home/anotheruser/", c, notify.InCloseWrite, notify.All); err != nil {
		log.Fatal(err)
	}*/
	defer notify.Stop(c)




	for ei := range c {

		switch ei.Event() {
		case notify.Create:
			Info.Println(ei.Event(), ei.Path())
		case notify.InCloseWrite:
			Info.Println(ei.Event(), ei.Path())
			if ValidateDeb(ei.Path()){
				debfile := moveDebs(ei.Path(), repotmp)
				CallAptlyAdd(debfile)
				CallAptlyShow()
				CallAptlyPublish()
			}

		}

	}

}
