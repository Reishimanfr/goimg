> [!WARNING]
> This readme is nowhere near finished.

## Goimg - A simple file hosting service
Sick of the absurdly low upload limits on discord? Want to upload your game replays without having to relay on 3rd party services? This program may be for you!
<br><br>Goimg is a very simple HTTP(s) server that allows you to embed large files on external websites like discord or many more. Everything is hosted locally so you know your files are safe!

## Features
- Database-directory sync (if you delete files from the files folder they'll also be deleted from the database automatically)
- Caching
- Video streaming
- Lots of customizability
- Easy integration with thunar and ShareX (soon!)

## How to self-host
### 1. Building the binary
> [!WARNING]
> This assumes you have everything downloaded and working correctly

> [!WARNING]
> If you're using the `-safe` flag you may need to run the binary as root
```sh
git clone --depth=1 https://github.com/Reishimanfr/goimg
cd goimg
go build
```

### 2. Downloading a pre-built binary
Simply go to the releases page and choose your binary

## Making the API accessible outside your local network
If you don't have a custom domain I highly recommend you get one from https://noip.com. They're completely free and you can configure them to auto-update the IP address if you can't set a static IP on your router. (Not sposored but I WISH noip please reach out :3)

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
Don't to forget to run it as root!
### 2. HTTP setup (not secure)
Simply run the binary. Yup, that's it.<br>
You can alternatively change the port used by the server with the `-port` flag

## Flags
| Name               | Description                                                 | Default value |
| ----               | -----------                                                 | :-----------: |
| `-max-file-size`   | Sets the max allowed file size (in MB)                      |  50           |
| `-storage-type`    | Sets where files should be saved                            | on-disk       |
| `-token-size-bits` | Length of the generated tokens (more = secure but annoying) | 64            |
| `-dev`             | Enables debugging stuff                                     | false         |
| `-port`            | Sets what port the server runs on. Ignored with `-secure`   | 8080          |
| `-secure`          | Enables HTTPS                                               | false         |


## Integrations
### Thunar
#### 1. Create a script to upload images 
```bash
#!/bin/bash

FILE="$1"
TOKEN="$2"
URL="[your upload url here]"             # For example nullptr.ddns.net/upload
BASE_URL="[your general url here]"       # For example nullptr.ddns.net         (notice the lack of /upload here)

response=$(curl -s -w "%{http_code}" -o /tmp/upload_response.json -X POST -F "file=@${FILE}" -H "Authorization: ${TOKEN}" ${URL})

if [ "$response" -eq 200 ]; then
    filename=$(jq -r '.key' /tmp/upload_response.json)

    if [ "$filename" != "null" ] && [ -n "$filename" ]; then
        upload_url="${BASE_URL}${filename}"

        # WARNING: This was tested on wayland and worked fine, may not work well for u if ur on x11 idk        
        echo -n "$upload_url" | xclip -selection clipboard

        # Success notification
        # If u don't want this just comment the line out
        notify-send "Upload Successful" "The file has been successfully uploaded.\nURL copied to clipboard: $upload_url"
    else
        # Failure if the filename (key) is missing or invalid
        # Same here but I suggest you leave this
        notify-send "Upload Failed" "Failed to extract filename from the server response."
    fi
else
    notify-send "Upload Failed" "The file upload failed with status code $response."
    echo "$response"
fi
```
#### 2. Add a new custom action in thunar
Open thunar, click `Edit > Configure custom actions...`<br>
Add a new action and name it whatever you like<br>
Set the command to `/path/to/your/upload/script %f [your access token the one that shows when you first launch the server]`<br>
Go to `Appearance Conditions` and enable whatever you want<br>

### ShareX
Soon :3
