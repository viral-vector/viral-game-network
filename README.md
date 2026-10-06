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
    B. Local registry
        * docker build -t localhost:5000/viral-game-network/bbserver:1.0.0 .
        * docker push localhost:5000/viral-game-network/bbserver:1.0.0

## License

Project code is licensed under the custom [Viral Game Network Use-Only License](LICENSE),
copyright (c) 2026 Viral Vector.

You may use and modify it for your own personal, organizational, or commercial use.
You may not redistribute the original or modified software, source code, binaries,
packages, or container images without prior written permission from Viral Vector,
even if you provide credit. You must retain notices and must not claim the original
work as your own. See LICENSE for the full terms.

This is a source-available license with redistribution restrictions.
Third-party dependencies and files carrying their own license notices remain subject
to those licenses, including the HashiCorp MPL-2.0 test template in bin/ioc/__tests__.
