package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type ConfApi struct {
	router     *mux.Router
	supervisor *Supervisor
}

// NewLogtail creates a Logtail object
func NewConfApi(supervisor *Supervisor) *ConfApi {
	return &ConfApi{router: mux.NewRouter(), supervisor: supervisor}
}

// CreateHandler creates http handlers to process the program stdout and stderr through http interface
func (ca *ConfApi) CreateHandler() http.Handler {
	ca.router.HandleFunc("/conf/supervisor_conf", ca.getSupervisorConfFile).Methods("GET")
	ca.router.HandleFunc("/conf/supervisor_conf", ca.saveSupervisorConfFile).Methods("POST")
	ca.router.HandleFunc("/conf/{program}", ca.getProgramConfFile).Methods("GET")
	ca.router.HandleFunc("/conf/{program}", ca.getProgramConfFile).Methods("POST")
	return ca.router
}

func (ca *ConfApi) getProgramConfFile(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	if vars == nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	programName := vars["program"]
	programConfigPath := getProgramConfigPath(programName, ca.supervisor)
	if programConfigPath == "" {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	b, err := readFile(programConfigPath)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(b)
}

type SaveProgramConfFileParams struct {
	Content string
}

func (ca *ConfApi) getSupervisorConfFile(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Write(bytes.NewBufferString(ca.supervisor.GetConfig().String()).Bytes())
}
func (ca *ConfApi) saveSupervisorConfFile(writer http.ResponseWriter, request *http.Request) {
	var p SaveProgramConfFileParams
	writer.WriteHeader(http.StatusOK)
	//writer.Write(bytes.NewBufferString(ca.supervisor.GetConfig().String()).Bytes())
	err := json.NewDecoder(request.Body).Decode(&p)
	path := ca.supervisor.GetConfig().GetConfigFile()
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		writer.WriteHeader(http.StatusForbidden)
		return
	}
	defer file.Close() // 关闭文件
	defer ca.supervisor.Reload(true)

	ioWriter := bufio.NewWriter(file)
	ioWriter.WriteString(p.Content)
	ioWriter.Flush()
	writer.WriteHeader(http.StatusOK)
}
func (ca *ConfApi) saveProgramConfFile(writer http.ResponseWriter, request *http.Request) {
	var p SaveProgramConfFileParams
	vars := mux.Vars(request)
	err := json.NewDecoder(request.Body).Decode(&p)

	if vars == nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	programName := vars["program"]
	programConfigPath := getProgramConfigPath(programName, ca.supervisor)
	if programConfigPath == "" {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	file, err := os.OpenFile(programConfigPath, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		writer.WriteHeader(http.StatusForbidden)
		return
	}

	defer file.Close() // 关闭文件
	defer ca.supervisor.Reload(true)
	ioWriter := bufio.NewWriter(file)
	ioWriter.WriteString(p.Content)
	ioWriter.Flush()
	writer.WriteHeader(http.StatusOK)

}
