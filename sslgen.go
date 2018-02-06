package main

//This script was created using the instructions at
//https://cybersyndicates.com/2016/12/egressing-bluecoat-with-cobaltstike-letsencrypt/

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func letsEncrypt(domain string) string {
	cmd := exec.Command("apt-get", "-y", "install", "openssl", "git", "apache2")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatalf("Problem installing openssl, git, apache2: %s", err)
	}
	cmd = exec.Command("service", "apache2", "start")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatalf("Problem starting apache2: %s", err)
	}
	cmd = exec.Command("apt-get", "update")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatalf("Problem updating via apt-get: %s", err)
	}
	if _, err := os.Stat("/opt/letsencrypt"); err != nil {
		if os.IsNotExist(err) {
			cmd = exec.Command("git", "clone", "https://github.com/certbot/certbot", "/opt/letsencrypt")
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			if err := cmd.Run(); err != nil {
				log.Fatalf("Problem git cloning certbot: %s", err)
			}
		} else {
			log.Fatalf("Problem checking if letsencrypt directory exists: %s", err)
		}
	}
	cmd = exec.Command("/opt/letsencrypt/letsencrypt-auto", "--apache", "-d", domain, "-n", "--register-unsafely-without-email", "--agree-tos")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatalf("Problem installing ssl certificate: %s", err)
	}
	cmd = exec.Command("service", "apache2", "stop")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Printf("Problem stopping apache2: %s", err)
	}
	return "/etc/letsencrypt/live/" + domain
}

func genCSKeystore(csdir string, domain string, password string) {
	letsencryptDir := letsEncrypt(domain)

	if !checkCertDeps() {
		log.Fatal("Check to make sure keytool and openssl are installed and in PATH")
	}
	cmd := exec.Command("openssl", "pkcs12", "-export", "-in", letsencryptDir+"/fullchain.pem", "-inkey",
		letsencryptDir+"/privkey.pem", "-out", letsencryptDir+"/"+domain+".pkcs", "-name", domain,
		"-passout", "pass:"+password)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatalf("Problem generating PKCS12 certificate: %s", err)
	}
	cmd = exec.Command("keytool", "-importkeystore", "-deststorepass", password, "-destkeypass", password,
		"-destkeystore", letsencryptDir+"/"+domain+".store", "-srckeystore", letsencryptDir+"/"+domain+".pkcs",
		"-srcstoretype", "PKCS12", "-srcstorepass", password, "-alias", domain)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatalf("Problem creating Java keystore: %s", err)
	}
	cmd = exec.Command("cp", letsencryptDir+"/"+domain+".store", csdir)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatalf("Problem copying keystore to cobalt strike directory: %s", err)
	}
}

//checkCertDeps checks for the existance of the keytool and openssl utilities
func checkCertDeps() bool {
	output, err := exec.Command("which", "keytool").Output()
	if err != nil {
		log.Fatal("Problem checking installation of: keytool")
	}
	if len(output) < 4 {
		return false
	}
	output, err = exec.Command("which", "openssl").Output()
	if err != nil {
		log.Fatal("Problem checking installation of: openssl")
	}
	if len(output) < 4 {
		return false
	}
	return true
}

func main() {
	genCSKeystore("/root", "swansonmedical.com", "Mike6192")
	fmt.Println("Your Cobalt Strike keystore has been created and placed in your Cobalt Strike folder. You will need to place it in the same directory as your C2 profile for it to work.")
}
