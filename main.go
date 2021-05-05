package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/julienschmidt/httprouter"
	uuid "github.com/satori/go.uuid"
)

const version = "1.2.0"

var mutex = &sync.Mutex{}
var conf Config
var secretMap = make(map[string]Secret)

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

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
		http.Redirect(w, r, redirectTarget, 302)
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
	for key, value := range secretMap {
		url := key
		count := value.counter
		name := value.name
		test := template.URL("/secret/" + url)
		fmt.Println(test)
		ActiveLinks = append(ActiveLinks, ActiveLink{GetTypeToString(value.ofType), test, count, name})
	}
	tmpl := template.Must(template.ParseFiles("html/active.html"))
	data := pageData{ActiveLinks, conf.Logo, template.HTML(BuildFooter(conf.Privacy, conf.Mail))}
	tmpl.Execute(w, data)
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
	link := ps.ByName("link")
	tmpl := template.Must(template.ParseFiles("html/getsecret.html"))
	isPas := isPasswordProtected(link)
	fmt.Println("Testing Password " + strconv.FormatBool(isPas))
	data := secretGetPageData{link, conf.Logo, template.HTML(BuildFooter(conf.Privacy, conf.Mail)), template.HTML(BuildPasswordInput(isPas))}
	tmpl.Execute(w, data)
}

/*isPasswordProtected ...*/
func isPasswordProtected(link string) bool {
	secretData, oks := secretMap[link]
	if oks {
		if len(secretData.pass) > 0 {
			return true
		}
		return false
	}
	return false
}

/*GetSecret ...*/
func GetSecret(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	secretLink := r.FormValue("GetSecret")
	pass := r.FormValue("password")

	fmt.Println("Testing Password " + pass)

	mutex.Lock()
	defer mutex.Unlock()

	secretData, oks := secretMap[secretLink]
	tmpl := template.Must(template.ParseFiles("html/secret.html"))
	if oks {

		if isPasswordProtected(secretLink) == true {
			if pass != secretData.pass {
				fmt.Println("WRONG Password " + pass)
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
		newCounter := secretData.counter - 1
		secretMap[secretLink] = Secret{secretData.data, secretData.ofType, newCounter, secretData.name, secretData.pass}
		if newCounter < 1 {
			delete(secretMap, secretLink)
			if secretData.ofType == File {
				DeleteFileIfExists(secretLink + BytesToString(secretData.data))
			}
		}
	} else {
		GetGonePage(w, r, ps, false)
	}
}

/*PostTextSecret ...*/
func PostTextSecret(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	validateAdmin(w, r, ps)

	//w.Header().Set("Content-Type", "text/plain")
	u1 := uuid.Must(uuid.NewV4()).String()
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Could not read input!")
	} else {
		sec := r.FormValue("secret")
		cou := r.FormValue("count")
		name := r.FormValue("name")
		pass := r.FormValue("password")

		fmt.Println("creating with password " + pass)

		c, cerr := strconv.Atoi(cou)
		if cerr != nil {
			c = 1
		}
		secretMap[u1] = Secret{[]byte(sec), Text, c, name, pass}
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
	delete(secretMap, link)
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
		c, cerr := strconv.Atoi(cou)
		if cerr != nil {
			c = 1
		}
		secretMap[u1] = Secret{[]byte(handler.Filename), File, c, name, pass}

		tmpl := template.Must(template.ParseFiles("html/preview.html"))
		fileName := template.HTML("File Uploaded!")
		data := secretPreviewData{fileName, (conf.Url + "/secret/" + u1), conf.Logo,
			template.HTML(BuildFooter(conf.Privacy, conf.Mail))}
		tmpl.Execute(w, data)
		io.WriteString(w, conf.Url+"/secret/"+u1)
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
	http.Redirect(w, r, redirectTarget, 302)
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
	http.Redirect(w, r, "/login", 302)
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

	router.DELETE("/secret/:link", Delete)
	http.ListenAndServe(":8080", router)
}
