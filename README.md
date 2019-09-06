# OneTimeLink

Easily share passwords, text, and files between people with an easy link for them to visit. Each secret that is created can be limited to a number of visits and all data will be deleted after this limit is reached.

Special Thanks to [Christoph](https://github.com/raytracer) who laid the groundwork for this project


## Build

To compile the application GoLang needs to be installed on the system and then call

```bash
env go build *.go
```

## To Run

to start the Program just call the following

```bash
./main -p <YOURPASSWORD>
```

this will startup the go application on localhost:8080. To login just use as User "admin" and your password.

Other options to customize the application are also available

- -u	URL that the application is running behind "secret.myURL.com"
- -g    URL to Data-privacy policy, (shown in the footer)
- -m  E-Mail address to contact the administrator (shown in the footer) 
- -l     URL to custom logo to display (SVG is recommended)


## Docker

To run OneTimeLink in a Docker container first build the application then build the docker image 

```bash
docker build -t onetimelink .
```

To run just call the following replacing <..> with your variables.

```
docker run -e "PASSWORD=<YourPassword>" -d -t -p 9001:8080 --name onetimelink onetimelink:latest
```

Example with  multiple variables

```
docker run -e "PASSWORD=<YourPassword>" -e "LOGO=<LinkToCustomLOGO>" -e "URL=<secret.MyURL.com>" -d -t -p 9001:8080 --name onetimelink onetimelink:latest
```
