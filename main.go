package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/kevinburke/ssh_config"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

func doHOST(wg *sync.WaitGroup, q chan string) {
	defer wg.Done()
	for {
		hosts, ok := <-q
		if !ok {
			return
		}
		fmt.Println("ssh: ", hosts)
		host := hosts
		user := ssh_config.Get(host, "User")
		addr := ssh_config.Get(host, "Hostname") + ":" + ssh_config.Get(host, "Port")
		auth := []ssh.AuthMethod{}
		buf, err := ioutil.ReadFile(ssh_config.Get(host, "IdentityFile"))
		if err != nil {
			log.Println(err)
		}
		key, err := ssh.ParsePrivateKey(buf)
		if err != nil {
			log.Println(err)
		}
		auth = append(auth, ssh.PublicKeys(key))
		config := &ssh.ClientConfig{
			User:            user,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Auth:            auth,
		}
		conn, err := ssh.Dial("tcp", addr, config)
		if err != nil {
			log.Println(err)
		}
		defer conn.Close()

		session, err := conn.NewSession()
		if err != nil {
			log.Println(err)
		}
		defer session.Close()
		//Check whoami
		var b bytes.Buffer
		session.Stdout = &b
		remote_command := "/usr/bin/whoami"
		if err := session.Run(remote_command); err != nil {
			log.Fatal("Failed to run: " + err.Error())
		}
		log.Println(remote_command + ":" + b.String())
	}
}

func main() {
	var wg sync.WaitGroup
	q := make(chan string, 10)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go doHOST(&wg, q)
	}
	filename := "list.conf"
	// ファイルオープン
	fp, err := os.Open(filename)
	if err != nil {
		// エラー処理
	}
	defer fp.Close()
	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		q <- scanner.Text()
	}
	if err = scanner.Err(); err != nil {
		// エラー処理
		log.Println(err)
	}
	close(q)
	wg.Wait()
}
