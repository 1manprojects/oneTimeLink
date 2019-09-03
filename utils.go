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
		"<a class=\"footer-link\" href=\"https://1manprojects.de\">Version " + version + "</a>" +
		//"<img class=\"gitlab-logo\" src=\"../static/GitHub-Mark-64px.png\"></img>" +
		"</div>"
	return footer
}
