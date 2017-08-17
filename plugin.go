package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/Sirupsen/logrus"
)

// Plugin holds the plugin information
type Plugin struct {
	Host       string
	Login      string
	Password   string
	Key        string
	Name       string
	Version    string
	Sources    string
	Inclusions string
	Exclusions string
	Language   string
	Profile    string
	Encoding   string
	LcovPath   string
	Debug      bool

	Path      string
	Repo      string
	Branch    string
	BranchOut string
	Default   string // default master branch
}

// Exec runs the plugin
func (p *Plugin) Exec() error {

	logrus.Println("Executing sonar analysis")

	if err := p.buildRunnerProperties(); err != nil {
		logrus.Println(err)
		return err
	}

	if err := p.execSonarRunner(); err != nil {
		logrus.Println(err)
		return err
	}

	p.writePipelineLetter()
	return nil
}

func (p Plugin) buildRunnerProperties() error {

	p.Key = strings.Replace(p.Key, "/", ":", -1)

	tmpl, err := template.ParseFiles("/opt/sonar/conf/sonar-runner.properties.tmpl")
	if err != nil {
		return err
	}

	f, err := os.Create("/opt/sonar/conf/sonar-runner.properties")
	defer f.Close()
	if err != nil {
		logrus.Println("Error creating file!")
		return err
	}

	if p.Debug {
		err = tmpl.ExecuteTemplate(os.Stdout, "sonar-runner.properties.tmpl", p)
		if err != nil {
			return err
		}
	}

	err = tmpl.ExecuteTemplate(f, "sonar-runner.properties.tmpl", p)
	if err != nil {
		return err
	}

	return nil
}

func (p Plugin) execSonarRunner() error {
	// run archive command
	cmd := exec.Command("java", "-jar", "/opt/sonar/runner.jar", "-Drunner.home=/opt/sonar/")
	printCommand(cmd)
	output, err := cmd.CombinedOutput()
	printOutput(output)

	if err != nil {
		return err
	}

	return nil
}

func (p Plugin) writePipelineLetter() {

	f, err := os.OpenFile(".Pipeline-Letter", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		logrus.Println("!!> Error creating / appending to .Pipeline-Letter")
		return
	}
	defer f.Close()

	if _, err = f.WriteString(fmt.Sprintf("*SONAR*: %s/dashboard/index/%s\n", p.Host, strings.Replace(p.Key, "/", ":", -1))); err != nil {
		logrus.Println("!!> Error writing to .Pipeline-Letter")
	}
}

func printCommand(cmd *exec.Cmd) {
	logrus.Printf("==> Executing: %s\n", strings.Join(cmd.Args, " "))
}

func printOutput(outs []byte) {
	if len(outs) > 0 {
		logrus.Printf("==> Output: %s\n", string(outs))
	}
}
