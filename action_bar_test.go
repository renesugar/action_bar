package action_bar

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/test/utils"

	"github.com/headzoo/surf"
)

var Server *httptest.Server
var CurrentUser *User
var actionBar *ActionBar

type User struct {
	Name string
	Role string
}

func (user *User) DisplayName() string {
	return user.Name
}

type AdminAuth struct {
}

func (AdminAuth) LoginURL(c *admin.Context) string {
	return "/auth/login"
}

func (AdminAuth) LogoutURL(c *admin.Context) string {
	return "/auth/logout"
}

func (AdminAuth) GetCurrentUser(c *admin.Context) qor.CurrentUser {
	if CurrentUser.Role != "admin" {
		return nil
	}
	return CurrentUser
}

// Init
func init() {
	Root, _ := os.Getwd()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		tmpl, err := template.New("home.tmpl").Funcs(actionBar.FuncMap(w, req)).ParseFiles(Root + "/test/views/home.tmpl")
		if err != nil {
			fmt.Printf("Execute template have error %v\n", err)
		}
		tmpl.Execute(w, struct {
			ActionBar   template.HTML
			CurrentUser *User
		}{
			ActionBar:   actionBar.Render(w, req),
			CurrentUser: CurrentUser,
		})
	})
	db := utils.TestDB()
	Server = httptest.NewServer(mux)
	Admin := admin.New(&qor.Config{DB: db})
	Admin.MountTo("/admin", mux)
	Admin.AddResource(User{})

	actionBar = New(Admin, AdminAuth{})
}

func TestAuth(t *testing.T) {
	bow := surf.NewBrowser()
	CurrentUser = &User{Name: "Kaea", Role: ""}
	bow.Open(Server.URL + "/")
	if !strings.Contains(bow.Body(), "LOGIN") {
		t.Errorf(color.RedString("Should Get `LOGIN` link if current user is normal user"))
	}

	CurrentUser = &User{Name: "Admin", Role: "admin"}
	bow.Open(Server.URL + "/")
	if !strings.Contains(bow.Body(), "LOGOUT") {
		t.Errorf(color.RedString("Should Get `LOGOUT` link if current user is admin user"))
	}
}

func TestEditMode(t *testing.T) {
	bow := surf.NewBrowser()

	// Default doesn't checked checkbox
	bow.Open(Server.URL + "/")
	html, _ := bow.Find(".qor-actionbar__module").Html()
	if strings.Contains(html, "checked") {
		t.Errorf(color.RedString("Should don't checked checkbox"))
	}

	// Should checked checkbox after switch mode
	bow.Open(Server.URL + "/admin/action_bar/switch_mode?checked=true")
	html, _ = bow.Find(".qor-actionbar__module").Html()
	if !strings.Contains(html, "checked") {
		t.Errorf(color.RedString("Should checked checkbox"))
	}
}

func TestInilneMode(t *testing.T) {
	bow := surf.NewBrowser()

	// Default should not have inline edit button
	bow.Open(Server.URL + "/")
	if bow.Find(".qor-actionbar-button").Length() != 0 {
		t.Errorf(color.RedString("Should don't have inline edit button"))
	}

	// Should have inline edit button if in edit mode
	bow.Open(Server.URL + "/admin/action_bar/switch_mode?checked=true")
	if bow.Find(".qor-actionbar-button").Length() == 0 {
		t.Errorf(color.RedString("Should have inline edit button"))
	}
}
