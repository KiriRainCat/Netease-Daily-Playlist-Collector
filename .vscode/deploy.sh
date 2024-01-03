#!/bin/bash

go-winres make --in ".vscode/winres/winres.json"

go build -ldflags "-w -s" -o 网易云曲目收集器.exe