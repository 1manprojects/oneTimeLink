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
func BuildPasswordInput(protected bool, twoFa string) string {
	if protected {
		return "<label class=\"pass-label\" for=\"password\">Password is required for this secret</label>" +
			"<input type=\"password\" autocomplete=\"off\" autocorrect=\"off\" autocapitalize=\"off\" spellcheck=\"false\" id=\"password\" name=\"password\">"
	}
	if len(twoFa) > 0 {
		id := []rune(twoFa)
		return "<span>Please provide the displayed text to the person who provided you with this link. You will be automatically forwarded to your data after confirmation. </span>" +
			"<label class=\"two-Fa-Check\" id=\"two-Fa-CheckLabel\">" + string(id[0:5]) + "</label>" +
			"<input id=\"password\" name=\"password\" type=\"hidden\" value=\"" + twoFa + "\"></input>" +
			"<span class=\"warning_span\">Do not reload this page or close the window!</span>"
	}
	return ""
}

/*GetFailurMessage ...*/
func GetFailurMessage(protected bool) string {
	if !protected {
		return "The Information you are trying to access does no longer exists. Either the link is invalid or the Information has already been retrieved. If you have received this link and see this page please contact the person who provided you with this link to have a new one sent to you."
	}
	return "You have entered an invalid Password"
}
