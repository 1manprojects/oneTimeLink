package main

import (
	"os"
)

/*FileExits ...*/
func FileExits(id string) bool {
	exists := false
	if _, err := os.Stat("./uploads/" + id); !os.IsNotExist(err) {
		exists = true
	}
	return exists
}

/*BytesToString ...*/
func BytesToString(data []byte) string {
	return string(data[:])
}

/*DeleteFileIfExists ...*/
func DeleteFileIfExists(id string) error {
	if _, err := os.Stat("./uploads/" + id); !os.IsNotExist(err) {
		os.Remove("./uploads/" + id)
		if err != nil {
			return err
		}
	}
	return nil
}

/*GetTypeToString ...*/
func GetTypeToString(i SecretType) string {
	if i == File {
		return "File"
	}
	if i == Text {
		return "Text"
	}
	return "ND"
}

/*BuildFooter ...*/
func BuildFooter(privacy string, mailto string) string {
	footer := "<div class=\"footer-info\"'>"
	if privacy != "" {
		footer += "<a class=\"footer-link\" href=\"" + conf.Privacy + "\">Privacy Policy</a>"
	}
	if mailto != "" {
		footer += "<a class=\"footer-link\" href=\"mailto:" + conf.Mail + "\">Contact</a>"
	}
	footer += "</div><div class=\"version-info\">" +
		"<a class=\"footer-link\" href=\"https://github.com/1manprojects/oneTimeLink\">Version " + version + "</a>" +
		"<img class=\"gitlab-logo\" src=\"../static/GitHub-Mark-64px.png\"></img>" +
		"</div>"
	return footer
}

/*BuildPasswordInput ...*/
func BuildPasswordInput(protected bool) string {
	if protected == true {
		return "<label for=\"password\">Password is required for this secret</label>" +
			"<input type=\"password\" autocomplete=\"off\" autocorrect=\"off\" autocapitalize=\"off\" spellcheck=\"false\" id=\"password\" name=\"password\">"
	}
	return ""
}

/*GetFailurMessage ...*/
func GetFailurMessage(protected bool) string {
	if protected == false {
		return "The Information you are trying to access does no longer exists. Either the link is invalid or the Information has alread been retrived. If you have recived this link and see this page please contact the person who proviede you with this link to have a new one sent to you."
	}
	return "You have enterd a invalid Password"
}
