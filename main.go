package main

import (
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/julienschmidt/httprouter"
	uuid "github.com/satori/go.uuid"
)

const version = "1.2.0"

var conf Config
var secretMap = NewSecretMap()

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

type SecretMap struct {
	sync.RWMutex
	secrets map[string]Secret
}

func NewSecretMap() *SecretMap {
	return &SecretMap{
		secrets: make(map[string]Secret),
	}
}

/*CleanUp ...*/
func CleanUp() {
	for {
		time.Sleep(30000 * time.Millisecond)
		elements := []string{}
		secretMap.RLock()
		for key, value := range secretMap.secrets {
			if value.validFor > -1 {
				if time.Since(value.createdOn) > time.Duration(int(time.Minute)*value.validFor) {
					elements = append(elements, key)
				}
			}
		}
		secretMap.RUnlock()
		if len(elements) > 0 {
			secretMap.Lock()
			for _, id := range elements {
				delete(secretMap.secrets, id)
			}
			secretMap.Unlock()
		}
	}
}

/*GetCreatePage ...*/
func GetCreatePage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	validateAdmin(w, r, ps)
	tmpl := template.Must(template.ParseFiles("html/create.html"))
	var ActiveLinks []ActiveLink
	data := pageData{ActiveLinks, conf.Logo, template.HTML(BuildFooter(conf.Privacy, conf.Mail))}
	tmpl.Execute(w, data)
}

/*Validtae User ...*/
func validateAdmin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if getUserName(r) != conf.User {
		redirectTarget := "/login"
		http.Redirect(w, r, redirectTarget, http.StatusFound)
	}
}

/*GetFilePage ...*/
func GetFilePage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	validateAdmin(w, r, ps)
	tmpl := template.Must(template.ParseFiles("html/file.html"))
	var ActiveLinks []ActiveLink
	data := pageData{ActiveLinks, conf.Logo, template.HTML(BuildFooter(conf.Privacy, conf.Mail))}
	tmpl.Execute(w, data)
}

/*GetActivePage ...*/
func GetActivePage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	validateAdmin(w, r, ps)
	var ActiveLinks []ActiveLink
	secretMap.RLock()
	for key, value := range secretMap.secrets {
		url := key
		count := value.counter
		name := value.name
		test := template.URL("/secret/" + url)
		validFor := "-"
		if value.validFor > -1 {
			tt := value.createdOn.Add(time.Duration(int(time.Minute) * value.validFor))
			minutesLeft := time.Now().Sub(tt).Minutes()
			validFor = strconv.Itoa(1 + (int(minutesLeft) * -1))
		}

		twoFa := "NONE"
		if len(value.twoFa) > 0 {
			twoFa = string([]rune(value.twoFa)[0:5])
		}
		if len(value.pass) > 0 {
			twoFa = "Pass"
		}
		state := getState(value)
		ActiveLinks = append(ActiveLinks, ActiveLink{GetTypeToString(value.ofType), test, count, name, twoFa, state, validFor})
	}
	secretMap.RUnlock()
	sort.Slice(ActiveLinks, func(i, j int) bool {
		return ActiveLinks[i].Name < ActiveLinks[j].Name
	})
	tmpl := template.Must(template.ParseFiles("html/active.html"))
	data := pageData{ActiveLinks, conf.Logo, template.HTML(BuildFooter(conf.Privacy, conf.Mail))}
	tmpl.Execute(w, data)
}

/*getState ...*/
func getState(s Secret) string {
	if s.isActive {
		if s.visited {
			return "reset"
		}
		return "active"
	} else {
		return "activate"
	}
}

func isActive(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	secretMap.RLock()
	var p idRequest
	res := new(boolResponse)
	res.Result = false
	err := json.NewDecoder(r.Body).Decode(&p)
	if err == nil {
		secretData, oks := secretMap.secrets[p.Id]
		if oks {
			if p.Tfa == secretData.twoFa {
				res.Result = secretData.isActive
			}
		}
	}
	secretMap.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

/*GetGonePage ...*/
func GetGonePage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, invalidPass bool) {
	tmpl := template.Must(template.ParseFiles("html/gone.html"))
	var ActiveLinks []ActiveLink
	data := invalidPageData{ActiveLinks, conf.Logo, template.HTML(BuildFooter(conf.Privacy, conf.Mail)), GetFailurMessage(invalidPass)}
	tmpl.Execute(w, data)
}

/*GetLoginPage ...*/
func GetLoginPage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	tmpl := template.Must(template.ParseFiles("html/login.html"))
	var ActiveLinks []ActiveLink
	data := pageData{ActiveLinks, conf.Logo, template.HTML(BuildFooter(conf.Privacy, conf.Mail))}
	tmpl.Execute(w, data)
}

/*LoadSecret ...*/
func LoadSecret(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Println("USER-AGENT: " + r.Header.Get("User-Agent"))
	agent := strings.ToLower(r.Header.Get("User-Agent"))
	if strings.Contains(agent, "telegram") || strings.Contains(agent, "whatsapp") || strings.Contains(agent, "Synapse ") {
		io.WriteString(w, "Go away!")
		return
	}

	link := ps.ByName("link")
	tmpl := template.Must(template.ParseFiles("html/getsecret.html"))
	isPas := isPasswordProtected(link)
	twoFa := getTowFaValue(link)
	data := secretGetPageData{link, twoFa, conf.Logo, template.HTML(BuildFooter(conf.Privacy, conf.Mail)), template.HTML(BuildPasswordInput(isPas, twoFa))}
	tmpl.Execute(w, data)
}

/*isPasswordProtected ...*/
func isPasswordProtected(link string) bool {
	secretMap.RLock()
	secretData, oks := secretMap.secrets[link]
	secretMap.RUnlock()
	if oks {
		return len(secretData.pass) > 0
	}
	return false
}

/*getTowFaValue ...*/
func getTowFaValue(link string) string {
	secretMap.Lock()
	secretData, oks := secretMap.secrets[link]
	if oks {
		if len(secretData.twoFa) > 0 {
			if !secretData.visited {
				newData := secretData
				newData.visited = true
				secretMap.secrets[link] = newData
				secretMap.Unlock()
				return secretData.twoFa
			} else {
				secretMap.Unlock()
				return uuid.Must(uuid.NewV4()).String()
			}
		} else {
			secretMap.Unlock()
			return ""
		}
	}
	secretMap.Unlock()
	return ""
}

/*EnableSecret ...*/
func EnableSecret(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	validateAdmin(w, r, ps)
	secretMap.Lock()
	secretLink := r.FormValue("GetSecret")
	id := strings.Replace(secretLink, "/secret/", "", 1)
	secretData, oks := secretMap.secrets[id]
	if oks {
		if len(secretData.twoFa) > 0 && secretData.isActive {
			u1 := uuid.Must(uuid.NewV4()).String()
			newData := secretData
			newData.visited = false
			secretMap.secrets[u1] = newData
			delete(secretMap.secrets, id)
		} else {
			newData := secretData
			newData.isActive = true
			secretMap.secrets[id] = newData
		}
	}
	secretMap.Unlock()
	GetActivePage(w, r, ps)
}

/*DeleteSecret ...*/
func DeleteSecret(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	validateAdmin(w, r, ps)
	secretMap.Lock()
	secretLink := r.FormValue("DelSecret")
	id := strings.Replace(secretLink, "/secret/", "", 1)
	delete(secretMap.secrets, id)
	secretMap.Unlock()
	GetActivePage(w, r, ps)
}

/*GetSecret ...*/
func GetSecret(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	secretLink := r.FormValue("GetSecret")
	pass := r.FormValue("password")
	secretMap.RLock()
	secretData, oks := secretMap.secrets[secretLink]
	secretMap.RUnlock()
	tmpl := template.Must(template.ParseFiles("html/secret.html"))
	if oks {
		if secretData.isActive {
			if isPasswordProtected(secretLink) {
				if pass != secretData.pass {
					GetGonePage(w, r, ps, true)
					return
				}
			}
			if len(secretData.twoFa) > 0 {
				if pass != secretData.twoFa {
					GetGonePage(w, r, ps, true)
					return
				}
			}

			if secretData.ofType == File {
				if FileExits(secretLink + BytesToString(secretData.data)) {
					w.Header().Set("Content-Disposition", "attachment; filename="+BytesToString(secretData.data))
					http.ServeFile(w, r, "./uploads/"+secretLink+BytesToString(secretData.data))
				} else {
					println("ERROR file does not exist")
				}
			}
			if secretData.ofType == Text {
				Secret := template.HTML(BytesToString(secretData.data))
				data := secretPageHTMLData{Secret, conf.Logo, template.HTML(BuildFooter(conf.Privacy, conf.Mail))}
				tmpl.Execute(w, data)
			}
			secretMap.Lock()
			newCounter := secretData.counter - 1
			secretMap.secrets[secretLink] = Secret{secretData.data, secretData.ofType, newCounter, secretData.name, secretData.pass, secretData.twoFa, secretData.isActive, secretData.visited, time.Now(), -1}
			if newCounter < 1 {
				delete(secretMap.secrets, secretLink)
				if secretData.ofType == File {
					DeleteFileIfExists(secretLink + BytesToString(secretData.data))
				}
			}
			secretMap.Unlock()
		} else {
			//todo change to inactive Page
			GetGonePage(w, r, ps, false)
		}
	} else {
		GetGonePage(w, r, ps, false)
	}
}

/*PostTextSecret ...*/
func PostTextSecret(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	validateAdmin(w, r, ps)
	u1 := uuid.Must(uuid.NewV4()).String()
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Could not read input!")
	} else {
		sec := r.FormValue("secret")
		cou := r.FormValue("count")
		name := r.FormValue("name")
		pass := r.FormValue("password")
		validFor := r.FormValue("validFor")
		towFe := r.FormValue("2fE") == "true"
		twoFaString := ""
		if towFe {
			twoFaString = uuid.Must(uuid.NewV4()).String()
		}

		v, verr := strconv.Atoi(validFor)
		if verr != nil || v == 0 {
			v = -1
		}

		c, cerr := strconv.Atoi(cou)
		//if no number suplied or twoFactor is enalbe then link is valid only once
		if cerr != nil || (len(twoFaString) > 0) {
			c = 1
		}
		secretMap.Lock()
		secretMap.secrets[u1] = Secret{[]byte(sec), Text, c, name, pass, twoFaString, (len(twoFaString) < 1), false, time.Now(), v}
		secretMap.Unlock()
		tmpl := template.Must(template.ParseFiles("html/preview.html"))
		secret := template.HTML(sec)
		data := secretPreviewData{secret, (conf.Url + "/secret/" + u1), conf.Logo,
			template.HTML(BuildFooter(conf.Privacy, conf.Mail))}
		tmpl.Execute(w, data)
	}

}

/*Delete ...*/
func Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "text/plain")
	link := ps.ByName("link")
	secretMap.Lock()
	delete(secretMap.secrets, link)
	secretMap.Unlock()
	io.WriteString(w, "ok")
}

/*Upload ...*/
func Upload(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	validateAdmin(w, r, ps)

	u1 := uuid.Must(uuid.NewV4()).String()
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("upload.gtpl")
		t.Execute(w, token)
	} else {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		f, err := os.OpenFile("./uploads/"+u1+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)

		name := r.FormValue("name")
		cou := r.FormValue("filecount")
		pass := r.FormValue("password")
		validFor := r.FormValue("validFor")
		towFe := r.FormValue("2fE") == "true"
		twoFaString := ""
		if towFe {
			twoFaString = uuid.Must(uuid.NewV4()).String()
		}

		v, verr := strconv.Atoi(validFor)
		if verr != nil || v == 0 {
			v = -1
		}

		c, cerr := strconv.Atoi(cou)
		if cerr != nil {
			c = 1
		}
		secretMap.Lock()
		secretMap.secrets[u1] = Secret{[]byte(handler.Filename), File, c, name, pass, twoFaString, (len(twoFaString) < 1), false, time.Now(), v}
		secretMap.Unlock()
		tmpl := template.Must(template.ParseFiles("html/preview.html"))
		fileName := template.HTML("File Uploaded!")
		data := secretPreviewData{fileName, (conf.Url + "/secret/" + u1), conf.Logo,
			template.HTML(BuildFooter(conf.Privacy, conf.Mail))}
		tmpl.Execute(w, data)
	}
}

/*loginHandler ...*/
func loginHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	//ReadConfig()
	name := r.FormValue("name")
	pass := r.FormValue("password")
	redirectTarget := "/login"

	if name == conf.User && pass == conf.Password {
		// .. check credentials ..
		setSession(name, w)
		redirectTarget = "/create"
	}
	http.Redirect(w, r, redirectTarget, http.StatusFound)
}

/*setSession ...*/
func setSession(userName string, w http.ResponseWriter) {
	value := map[string]string{
		"name": userName,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(w, cookie)
	}
}

/*getUserName ...*/
func getUserName(r *http.Request) (userName string) {
	if cookie, err := r.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			userName = cookieValue["name"]
		}
	}
	return userName
}

/*clearSession ...*/
func clearSession(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
}

/*Logout ...*/
func Logout(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	clearSession(w)
	http.Redirect(w, r, "/login", http.StatusFound)
}

func init() {
	conf = Config{"admin", "", "", "", "", ""}
}

func main() {

	flag.StringVar(&conf.Password, "p", "", "Password for User login, Required!")
	flag.StringVar(&conf.Url, "u", "127.0.0.1:8080", "URL the server is reachable from, Optional")
	flag.StringVar(&conf.Logo, "l", "../static/logo.svg", "Http link to custom Logo, Optional")
	flag.StringVar(&conf.Privacy, "g", "", "Http link to privacy policy for server, Optional")
	flag.StringVar(&conf.Mail, "m", "", "E-mail of Server-Admin or Responsible, Optional")
	flag.Parse()

	if conf.Logo == "" {
		conf.Logo = "../static/logo.svg"
	}

	if conf.Password == "" {
		log.Fatal("Exiting, no password was set. Please specify a Password with -p PASSWORD")
	}

	go CleanUp()

	router := httprouter.New()
	router.ServeFiles("/static/*filepath", http.Dir("static"))

	router.GET("/secret/:link", LoadSecret)
	router.POST("/loadsecret", GetSecret)
	router.GET("/create", GetCreatePage)
	router.GET("/file", GetFilePage)
	router.GET("/active", GetActivePage)
	router.GET("/login", GetLoginPage)
	router.POST("/login", loginHandler)
	router.GET("/", loginHandler)
	router.POST("/secretText", PostTextSecret)
	router.POST("/upload", Upload)
	router.POST("/logout", Logout)
	router.POST("/enablesecret", EnableSecret)
	router.POST("/deletesecret", DeleteSecret)
	router.POST("/isActive", isActive)

	router.DELETE("/secret/:link", Delete)
	http.ListenAndServe(":8080", router)

}
