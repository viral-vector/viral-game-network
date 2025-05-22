# README #

Viral-Network (Viral Games Network Backend) is a game backend server written in Golang.

### What is this repository for? ###

* Game Network Backend Server written in Golang 
* 1.0.0

## Operations & Commands
1. VGN Image to registry
    * docker login -u [email] -p YOUR_DO_API_TOKEN registry.digitalocean.com
    * docker build -f docker/Dockerfile -t registry.digitalocean.com/vgn-global-registry/vgn:1.0.0 . 
    * docker push registry.digitalocean.com/vgn-global-registry/vgn:1.0.0
2. App Image to registry
    * docker build -t localhost:5000/viral-game-network/bbserver:1.0.0 .
        * Builds your image and tags it with the registry address 
    * docker push localhost:5000/viral-game-network/bbserver:1.0.0
        * Push your image to the local registry
3. Gen New Migration files
    * New-Item -Path "$(Get-Date -Format 'ddMMyyyy').go" -ItemType File
        * Run this command from the migrations folder

