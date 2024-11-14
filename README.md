> [!WARNING]
> This readme is nowhere near finished.

## Goimg - A simple file embedded for discord
Sick of the absurdly low upload limits on discord? Want to upload your game replays without having to relay on 3rd party services? This program may be for you!
<br><br>Goimg is a very simple HTTP server that allows you to embed large files on external websites like discord or many more. Everything is hosted locally so you know your files are safe!

## How to self-host
### 1. Building the binary
> [!WARNING]
> This assumes you have everything downloaded and working correctly

> [!WARNING]
> If you're using the `-safe` flag you may need to run the binary as root (with `sudo`)
```sh
git clone --depth=1 https://github.com/Reishimanfr/goimg
cd goimg
go build
./goimg
```

### 2. Downloading a pre-built binary
Simply go to the releases page and choose your binary

## Making the API accessible outside your local network
If you don't have a custom domain I highly recommend you get one from https://noip.com. They're completely free and you can configure them to auto-update the IP address if you can't set a static IP on your router

### 1. HTTPS setup (recommended)
#### 1.1 Open ports to your local network
You'll need to go to your router's config panel and open ports `443` and `80`. Due to all routers being different I cannot help you with the specifics on how to open ports on your router. Google is your best friend here

#### 1.2 Install `certbot` (for SSL certificates)
Example for arch linux:
```sh
pacman -S certbot
```

#### 1.3 Request an SSL certificate
```sh
certbot certonly --standalone -d [your domain name here]
```
This command will create 2 files and print out their path. You'll need to provide the paths to these files with flags (`-ssl-cert-path` and `-ssl-key-path`)

#### 1.4 Run the API server
Example:
```sh
./goimg -secure=true -ssl-cert-path="/path/to/your/cert" -ssl-key-path="/path/to/your/certkey"
```

### 2. HTTP setup (not secure)
todo
