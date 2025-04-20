# README #

Viral-Network (Viral Games Network Backend) is a game backend server written in Golang.

### What is this repository for? ###

* Game Network Backend Server written in Golang 
* 1.0.0

## Commands
1. Image to registry
    * docker build -t localhost:5000/viral-game-network/bbserver:1.0.0 .
        * Builds your image and tags it with the registry address 
    * docker push localhost:5000/viral-game-network/bbserver:1.0.0
        * Push your image to the local registry
2. Gen New Migration files
    * New-Item -Path "$(Get-Date -Format 'ddMMyyyy').go" -ItemType File
        * Run this command from the migrations folder

