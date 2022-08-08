# Oversaw
A play on the app "Overseer". This is a tool that makes the use of Unraid(or some kind of docker server), Overseer, Plex watchlist, and the new Plex Discover feature to add shows/movies to your Plex libraries. The goal and usecase for this app is so I can see a show I want to watch using Plex's discover feature, add the show to my watchlist, and the show will be added to Overseer and subsequently requested/queued for download. Note: This does treat the watchlist as a queueing system. If you already use the watchlist to save shows to watch later, then this is not the right tool for you.

## Requirements
 - Overseer
 - Unraid (or a server that is running docker)
 - Overseer is connected to sonarr/radarr or some kind of movie/tv show management system.

## Authentication
For this feature we will need two sets of keys
Overseer API Key - https://docs.overseerr.dev/using-overseerr/settings
Plex Token - https://support.plex.tv/articles/204059436-finding-an-authentication-token-x-plex-token/

## Build and run locally
``` go build -o oversaw ```

``` oversaw start-oversaw --token <PLEX-TOKEN> --apikey <OVERSEER-API-KEY> --overseer-url <overseer url or ip (http://192.168.0.100:5055 or https://requests.myplex.com  no need for trailing slash)> ```

## Run in docker
``` docker pull 08231990/oversaw:latest ```

``` docker run -e URL=<OVERSEER-URL> -e APIKEY=<OVERSEER-API-KEY> -e TOKEN=<PLEX-API-TOKEN> 08231990/oversaw:latest ```

## Run in Unraid
- Go to Unraid -> Docker -> Add Container (bottom of your screen)
- Make sure advanced view is on (toggle is near the top right corner of your screen)
- Name: Oversaw
- Repository: docker.io/08231990/oversaw
- Extra Parameters: --restart unless-stopped
- At the bottom, click "Add another Path, Port, Variable, Label or Device" (We are adding 3 environment variables)
- In the Configuration screen
    - Config type: Variable
    - Name: URL
    - Key: URL
    - Value: overseer url or ip (http://192.168.0.100:5055 or https://requests.myplex.com)
    - Click Add
- Now do this again except add in APIKEY and TOKEN instead of URL, (make sure everything URL, APIKEY, TOKEN are all caps)

## Enjoy
Now you can add to different shows/movies to watchlist and watch it get requested in Overseer (and subsequently downloaded)

## NOTES
- (Bug) I use the Overseer API search and match by name and year. Sometimes, rarely, two shows or movies with the same name are released in the same year and this may confuse the app. This is why I made it open source... ALL PULL REQUESTS ARE WELCOME