package modules

import (
	"NovaUserbot/locales"
	"NovaUserbot/utils"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/amarnathcjd/gogram/telegram"
)

func init() {
	RegisterModule("Plugins", loadPluginsModule)
}

func installPlugin(m *telegram.NewMessage) error {
	url := strings.TrimSpace(m.Args())
	if url == "" {
		_, err := eOR(m, locales.Tr("plugins.install_usage"))
		return err
	}

	if strings.Contains(url, "github.com") && !strings.Contains(url, "raw.githubusercontent.com") {
		url = strings.Replace(url, "github.com", "raw.githubusercontent.com", 1)
		url = strings.Replace(url, "/blob/", "/", 1)
	}

	msg, _ := eOR(m, locales.Tr("plugins.install_progress"))

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		errMsg := "unknown error"
		if err != nil {
			errMsg = err.Error()
		} else {
			errMsg = fmt.Sprintf("HTTP %d", resp.StatusCode)
			resp.Body.Close()
		}
		_, err = msg.Edit(locales.Trf("plugins.install_fetch_error", errMsg))
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		_, err = msg.Edit(locales.Trf("plugins.install_fetch_error", err.Error()))
		return err
	}

	content := string(body)
	if !strings.Contains(content, "package modules") || !strings.Contains(content, "func init()") {
		_, err = msg.Edit(locales.Tr("plugins.install_invalid"))
		return err
	}

	parts := strings.Split(url, "/")
	fileName := parts[len(parts)-1]
	if !strings.HasSuffix(fileName, ".go") {
		fileName += ".go"
	}
	pluginName := strings.TrimSuffix(fileName, ".go")
	destPath := filepath.Join("modules", "plugin_"+fileName)

	if err := os.WriteFile(destPath, body, 0644); err != nil {
		_, err = msg.Edit(locales.Trf("plugins.install_write_error", err.Error()))
		return err
	}

	out, err := utils.RunCommand("go build -o main .")
	if err != nil {
		os.Remove(destPath)
		_, err = msg.Edit(locales.Trf("plugins.install_build_error", out))
		return err
	}

	pluginsJson, _ := Db.Get(context.Background(), "INSTALLED_PLUGINS").Result()
	plugins := make(map[string]string)
	if pluginsJson != "" {
		json.Unmarshal([]byte(pluginsJson), &plugins)
	}
	plugins[pluginName] = url
	updated, _ := json.Marshal(plugins)
	Db.Set(context.Background(), "INSTALLED_PLUGINS", updated, 0)

	_, _ = msg.Edit(locales.Trf("plugins.install_success", pluginName))

	restartBot()
	return nil
}

func uninstallPlugin(m *telegram.NewMessage) error {
	name := strings.TrimSpace(m.Args())
	if name == "" {
		_, err := eOR(m, locales.Tr("plugins.uninstall_usage"))
		return err
	}

	destPath := filepath.Join("modules", "plugin_"+name+".go")
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		_, err = eOR(m, locales.Trf("plugins.uninstall_not_found", name))
		return err
	}

	if err := os.Remove(destPath); err != nil {
		_, err = eOR(m, locales.Trf("plugins.uninstall_error", err.Error()))
		return err
	}

	utils.RunCommand("go build -o main .")

	pluginsJson, _ := Db.Get(context.Background(), "INSTALLED_PLUGINS").Result()
	plugins := make(map[string]string)
	if pluginsJson != "" {
		json.Unmarshal([]byte(pluginsJson), &plugins)
	}
	delete(plugins, name)
	updated, _ := json.Marshal(plugins)
	Db.Set(context.Background(), "INSTALLED_PLUGINS", updated, 0)

	_, _ = eOR(m, locales.Trf("plugins.uninstall_success", name))
	restartBot()
	return nil
}

func listPlugins(m *telegram.NewMessage) error {
	pluginsJson, _ := Db.Get(context.Background(), "INSTALLED_PLUGINS").Result()
	plugins := make(map[string]string)
	if pluginsJson != "" {
		json.Unmarshal([]byte(pluginsJson), &plugins)
	}

	if len(plugins) == 0 {
		_, err := eOR(m, locales.Tr("plugins.list_none"))
		return err
	}

	response := locales.Tr("plugins.list_header")
	for name := range plugins {
		response += fmt.Sprintf(locales.Tr("plugins.list_entry"), name)
	}
	_, err := eOR(m, response)
	return err
}

func restartBotCmd(m *telegram.NewMessage) error {
	_, _ = eOR(m, locales.Tr("system.restarting"))
	restartBot()
	return nil
}

func shutdownBot(m *telegram.NewMessage) error {
	_, _ = eOR(m, locales.Tr("system.shutting_down"))
	os.Exit(0)
	return nil
}

func restartBot() {
	binary, err := os.Executable()
	if err != nil {
		return
	}
	syscall.Exec(binary, os.Args, os.Environ())
}

func loadPluginsModule() {
	handlers := []*Handler{
		{ModuleName: "Plugins", Command: "install", Description: "Install a plugin from URL", Func: installPlugin, DisAllowSudos: true},
		{ModuleName: "Plugins", Command: "uninstall", Description: "Uninstall a plugin by name", Func: uninstallPlugin, DisAllowSudos: true},
		{ModuleName: "Plugins", Command: "plugins", Description: "List installed plugins", Func: listPlugins},
		{ModuleName: "Plugins", Command: "restart", Description: "Restart the userbot", Func: restartBotCmd, DisAllowSudos: true},
		{ModuleName: "Plugins", Command: "shutdown", Description: "Shutdown the userbot", Func: shutdownBot, DisAllowSudos: true},
	}
	AddHandlers(handlers, client)
}
