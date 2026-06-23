package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sshbox/internal/service"
	"sshbox/internal/storage"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// testManager creates a fresh manager for testing.
func testManager(t *testing.T) *service.ConnectionManager {
	t.Helper()
	dir := t.TempDir()
	store := storage.NewConfigStoreWithPath(filepath.Join(dir, "config.json"))
	masterKey := []byte("0123456789abcdef0123456789abcdef")
	return service.NewConnectionManager(store, masterKey)
}

// newRoot creates a fresh root command tree bound to the given manager.
func newRoot(mgr *service.ConnectionManager) *cobra.Command {
	cmd := &cobra.Command{Use: "sshbox", Short: "轻量级SSH连接管理工具"}

	addCmd := &cobra.Command{
		Use: "add <name>", Short: "添加SSH连接", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			host, _ := cmd.Flags().GetString("host")
			port, _ := cmd.Flags().GetInt("port")
			user, _ := cmd.Flags().GetString("user")
			password, _ := cmd.Flags().GetString("password")
			tagStr, _ := cmd.Flags().GetString("tag")
			notes, _ := cmd.Flags().GetString("notes")
			if host == "" { return fmt.Errorf("缺少必填参数 --host") }
			if user == "" { return fmt.Errorf("缺少必填参数 --user") }
			if password == "" { return fmt.Errorf("缺少必填参数 --password") }
			var tags []string
			if tagStr != "" {
				for _, t := range strings.Split(tagStr, ",") {
					t = strings.TrimSpace(t)
					if t != "" { tags = append(tags, t) }
				}
			}
			if err := mgr.AddConnection(name, host, port, user, password, tags, notes); err != nil { return err }
			cmd.Printf("✓ 连接 %s 已添加\n", name)
			return nil
		},
	}
	addCmd.Flags().StringP("host", "H", "", "主机地址（必填）")
	addCmd.Flags().IntP("port", "P", 22, "端口号")
	addCmd.Flags().StringP("user", "u", "", "登录用户名（必填）")
	addCmd.Flags().StringP("password", "p", "", "登录密码（必填）")
	addCmd.Flags().StringP("tag", "t", "", "标签，多个用逗号分隔")
	addCmd.Flags().StringP("notes", "n", "", "备注信息")

	listCmd := &cobra.Command{
		Use: "list", Short: "列出所有SSH连接",
		RunE: func(cmd *cobra.Command, args []string) error {
			tag, _ := cmd.Flags().GetString("tag")
			search, _ := cmd.Flags().GetString("search")
			format, _ := cmd.Flags().GetString("format")
			conns, err := mgr.ListConnections(tag, search)
			if err != nil { return err }
			if len(conns) == 0 { cmd.Println("没有找到连接"); return nil }
			switch format {
			case "json":
				data, _ := json.MarshalIndent(conns, "", "  ")
				cmd.Println(string(data))
			default:
				headers := []string{"名称", "主机", "端口", "用户", "标签"}
				var rows [][]string
				for _, c := range conns {
					rows = append(rows, []string{c.Name, c.Host, fmt.Sprintf("%d", c.Port), c.User, strings.Join(c.Tags, ",")})
				}
				printTable(cmd.OutOrStdout(), headers, rows)
			}
			return nil
		},
	}
	listCmd.Flags().String("tag", "", "按标签筛选")
	listCmd.Flags().String("search", "", "搜索关键词")
	listCmd.Flags().String("format", "table", "输出格式（table/json）")

	showCmd := &cobra.Command{
		Use: "show <name>", Short: "查看连接详情", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conn, err := mgr.GetConnection(args[0])
			if err != nil { return err }
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "名称：    %s\n", conn.Name)
			fmt.Fprintf(out, "主机：    %s\n", conn.Host)
			fmt.Fprintf(out, "端口：    %d\n", conn.Port)
			fmt.Fprintf(out, "用户：    %s\n", conn.User)
			fmt.Fprintf(out, "密码：    ********\n")
			fmt.Fprintf(out, "标签：    %s\n", strings.Join(conn.Tags, ", "))
			fmt.Fprintf(out, "备注：    %s\n", conn.Notes)
			fmt.Fprintf(out, "创建时间：%s\n", conn.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Fprintf(out, "更新时间：%s\n", conn.UpdatedAt.Format("2006-01-02 15:04:05"))
			return nil
		},
	}

	editCmd := &cobra.Command{
		Use: "edit <name>", Short: "编辑SSH连接", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			var newName, host, user, password *string; var port *int; var tags *[]string; var notes *string
			if cmd.Flags().Changed("name") { v, _ := cmd.Flags().GetString("name"); newName = &v }
			if cmd.Flags().Changed("host") { v, _ := cmd.Flags().GetString("host"); host = &v }
			if cmd.Flags().Changed("port") { v, _ := cmd.Flags().GetInt("port"); port = &v }
			if cmd.Flags().Changed("user") { v, _ := cmd.Flags().GetString("user"); user = &v }
			if cmd.Flags().Changed("password") { v, _ := cmd.Flags().GetString("password"); password = &v }
			if cmd.Flags().Changed("tag") { v, _ := cmd.Flags().GetString("tag"); p := strings.Split(v, ","); tags = &p }
			if cmd.Flags().Changed("notes") { v, _ := cmd.Flags().GetString("notes"); notes = &v }
			if err := mgr.UpdateConnection(name, newName, host, port, user, password, tags, notes); err != nil { return err }
			if newName != nil { cmd.Printf("✓ 连接 %s 已重命名为 %s\n", name, *newName) } else { cmd.Printf("✓ 连接 %s 已更新\n", name) }
			return nil
		},
	}
	editCmd.Flags().StringP("name", "N", "", "连接名称（重命名）")
	editCmd.Flags().StringP("host", "H", "", "主机地址")
	editCmd.Flags().IntP("port", "P", 22, "端口号")
	editCmd.Flags().StringP("user", "u", "", "登录用户名")
	editCmd.Flags().StringP("password", "p", "", "登录密码")
	editCmd.Flags().StringP("tag", "t", "", "标签（替换现有标签）")
	editCmd.Flags().StringP("notes", "n", "", "备注信息")

	rmCmd := &cobra.Command{
		Use: "rm <name>", Short: "删除SSH连接", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := mgr.DeleteConnection(args[0]); err != nil { return err }
			cmd.Printf("✓ 连接 %s 已删除\n", args[0])
			return nil
		},
	}

	tagsCmd := &cobra.Command{
		Use: "tags [tag]", Short: "列出标签及其连接", Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				tag := args[0]
				conns, err := mgr.ListConnections(tag, "")
				if err != nil { return err }
				if len(conns) == 0 { cmd.Printf("标签 %s 下没有连接\n", tag); return nil }
				cmd.Printf("标签 %s (%d 个连接):\n\n", tag, len(conns))
				headers := []string{"名称", "主机", "端口", "用户"}
				var rows [][]string
				for _, c := range conns { rows = append(rows, []string{c.Name, c.Host, fmt.Sprintf("%d", c.Port), c.User}) }
				printTable(cmd.OutOrStdout(), headers, rows)
				return nil
			}
			counts, err := mgr.ListTags()
			if err != nil { return err }
			if len(counts) == 0 { cmd.Println("没有标签"); return nil }
			keys := make([]string, 0, len(counts))
			for k := range counts { keys = append(keys, k) }
			sort.Strings(keys)
			headers := []string{"标签", "连接数"}
			var rows [][]string
			for _, k := range keys { rows = append(rows, []string{k, fmt.Sprintf("%d", counts[k])}) }
			printTable(cmd.OutOrStdout(), headers, rows)
			return nil
		},
	}

	exportCmd := &cobra.Command{
		Use: "export", Short: "导出配置",
		RunE: func(cmd *cobra.Command, args []string) error {
			output, _ := cmd.Flags().GetString("output")
			tag, _ := cmd.Flags().GetString("tag")
			data, err := mgr.ExportJSON(tag)
			if err != nil { return err }
			if output != "" {
				if err := os.WriteFile(output, data, 0600); err != nil { return fmt.Errorf("写入文件失败: %w", err) }
				cmd.Printf("✓ 配置已导出到 %s\n", output)
			} else {
				cmd.Println(string(data))
			}
			return nil
		},
	}
	exportCmd.Flags().String("output", "", "输出文件路径")
	exportCmd.Flags().String("tag", "", "只导出指定标签的连接")

	importCmd := &cobra.Command{
		Use: "import <file>", Short: "导入配置", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			merge, _ := cmd.Flags().GetBool("merge")
			count, err := mgr.ImportFromFile(args[0], merge)
			if err != nil { return err }
			cmd.Printf("✓ 成功导入 %d 个连接\n", count)
			return nil
		},
	}
	importCmd.Flags().Bool("merge", false, "合并模式（保留现有连接）")

	passwordCmd := &cobra.Command{
		Use: "password <name>", Short: "查看连接密码", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pass, err := mgr.GetDecryptedPassword(args[0])
			if err != nil { return err }
			fmt.Fprintln(cmd.OutOrStdout(), pass)
			return nil
		},
	}

	cmd.AddCommand(addCmd, listCmd, showCmd, editCmd, rmCmd, tagsCmd, exportCmd, importCmd, passwordCmd)
	return cmd
}

func run(t *testing.T, mgr *service.ConnectionManager, args ...string) (string, error) {
	t.Helper()
	root := newRoot(mgr)
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestAddCommand(t *testing.T) {
	mgr := testManager(t)
	out, err := run(t, mgr, "add", "test-srv", "--host", "1.1.1.1", "--user", "root", "--password", "pass123", "--tag", "prod,web", "--notes", "test server")
	if err != nil { t.Fatalf("add failed: %v", err) }
	if !strings.Contains(out, "✓ 连接 test-srv 已添加") { t.Errorf("unexpected output: %q", out) }
}

func TestAddMissingHost(t *testing.T) {
	mgr := testManager(t)
	_, err := run(t, mgr, "add", "test-srv", "--user", "root", "--password", "pass")
	if err == nil { t.Error("add without --host should fail") }
}

func TestAddMissingUser(t *testing.T) {
	mgr := testManager(t)
	_, err := run(t, mgr, "add", "test-srv", "--host", "1.1.1.1", "--password", "pass")
	if err == nil { t.Error("add without --user should fail") }
}

func TestAddMissingPassword(t *testing.T) {
	mgr := testManager(t)
	_, err := run(t, mgr, "add", "test-srv", "--host", "1.1.1.1", "--user", "root")
	if err == nil { t.Error("add without --password should fail") }
}

func TestListCommand(t *testing.T) {
	mgr := testManager(t)
	run(t, mgr, "add", "s1", "--host", "1.1.1.1", "--user", "root", "--password", "pass", "--tag", "prod")
	run(t, mgr, "add", "s2", "--host", "2.2.2.2", "--user", "root", "--password", "pass", "--tag", "dev")
	out, err := run(t, mgr, "list")
	if err != nil { t.Fatalf("list failed: %v", err) }
	if !strings.Contains(out, "s1") || !strings.Contains(out, "s2") { t.Errorf("list output missing connections: %q", out) }
}

func TestListByTag(t *testing.T) {
	mgr := testManager(t)
	run(t, mgr, "add", "s1", "--host", "1.1.1.1", "--user", "root", "--password", "pass", "--tag", "prod")
	run(t, mgr, "add", "s2", "--host", "2.2.2.2", "--user", "root", "--password", "pass", "--tag", "dev")
	out, _ := run(t, mgr, "list", "--tag", "prod")
	if !strings.Contains(out, "s1") || strings.Contains(out, "s2") { t.Errorf("tag filter not working: %q", out) }
}

func TestListJSON(t *testing.T) {
	mgr := testManager(t)
	run(t, mgr, "add", "s1", "--host", "1.1.1.1", "--user", "root", "--password", "pass")
	out, _ := run(t, mgr, "list", "--format", "json")
	if !strings.Contains(out, "\"name\": \"s1\"") { t.Errorf("JSON output missing data: %q", out) }
}

func TestShowCommand(t *testing.T) {
	mgr := testManager(t)
	run(t, mgr, "add", "s1", "--host", "1.1.1.1", "--port", "2222", "--user", "admin", "--password", "secret", "--tag", "prod", "--notes", "my server")
	out, err := run(t, mgr, "show", "s1")
	if err != nil { t.Fatalf("show failed: %v", err) }
	if !strings.Contains(out, "1.1.1.1") { t.Errorf("show output missing host: %q", out) }
	if !strings.Contains(out, "2222") { t.Errorf("show output missing port: %q", out) }
	if !strings.Contains(out, "admin") { t.Errorf("show output missing user: %q", out) }
	if !strings.Contains(out, "********") { t.Errorf("show should mask password: %q", out) }
	if strings.Contains(out, "secret") { t.Errorf("show should not reveal password: %q", out) }
}

func TestEditCommand(t *testing.T) {
	mgr := testManager(t)
	run(t, mgr, "add", "s1", "--host", "1.1.1.1", "--user", "root", "--password", "pass")
	out, err := run(t, mgr, "edit", "s1", "--host", "2.2.2.2")
	if err != nil { t.Fatalf("edit failed: %v", err) }
	if !strings.Contains(out, "✓ 连接 s1 已更新") { t.Errorf("unexpected output: %q", out) }
	showOut, _ := run(t, mgr, "show", "s1")
	if !strings.Contains(showOut, "2.2.2.2") { t.Errorf("edit did not update host: %q", showOut) }
}

func TestEditShortFlags(t *testing.T) {
	mgr := testManager(t)
	run(t, mgr, "add", "s1", "--host", "1.1.1.1", "--user", "root", "--password", "pass")
	out, err := run(t, mgr, "edit", "s1", "-H", "3.3.3.3", "-u", "admin", "-p", "newpass")
	if err != nil { t.Fatalf("edit with short flags failed: %v", err) }
	if !strings.Contains(out, "✓ 连接 s1 已更新") { t.Errorf("unexpected output: %q", out) }
	showOut, _ := run(t, mgr, "show", "s1")
	if !strings.Contains(showOut, "3.3.3.3") { t.Errorf("-H not working: %q", showOut) }
	if !strings.Contains(showOut, "admin") { t.Errorf("-u not working: %q", showOut) }
}

func TestEditRename(t *testing.T) {
	mgr := testManager(t)
	run(t, mgr, "add", "s1", "--host", "1.1.1.1", "--user", "root", "--password", "pass")
	out, err := run(t, mgr, "edit", "s1", "--name", "renamed")
	if err != nil { t.Fatalf("edit rename failed: %v", err) }
	if !strings.Contains(out, "已重命名为 renamed") { t.Errorf("unexpected output: %q", out) }
	showOut, _ := run(t, mgr, "show", "renamed")
	if !strings.Contains(showOut, "1.1.1.1") { t.Errorf("rename lost data: %q", showOut) }
	// 旧名称应不存在
	_, err = run(t, mgr, "show", "s1")
	if err == nil { t.Error("old name should not exist after rename") }
}

func TestTagsCommand(t *testing.T) {
	mgr := testManager(t)
	run(t, mgr, "add", "s1", "--host", "1.1.1.1", "--user", "root", "--password", "pass", "--tag", "prod,web")
	run(t, mgr, "add", "s2", "--host", "2.2.2.2", "--user", "root", "--password", "pass", "--tag", "prod")
	out, err := run(t, mgr, "tags")
	if err != nil { t.Fatalf("tags failed: %v", err) }
	if !strings.Contains(out, "prod") || !strings.Contains(out, "2") { t.Errorf("tags output incorrect: %q", out) }
}

func TestExportCommand(t *testing.T) {
	mgr := testManager(t)
	run(t, mgr, "add", "s1", "--host", "1.1.1.1", "--user", "root", "--password", "pass")
	out, err := run(t, mgr, "export")
	if err != nil { t.Fatalf("export failed: %v", err) }
	if !strings.Contains(out, "s1") || !strings.Contains(out, "1.1.1.1") { t.Errorf("export output missing data: %q", out) }
}

func TestExportToFile(t *testing.T) {
	mgr := testManager(t)
	run(t, mgr, "add", "s1", "--host", "1.1.1.1", "--user", "root", "--password", "pass")
	tmpFile := filepath.Join(t.TempDir(), "export.json")
	out, err := run(t, mgr, "export", "--output", tmpFile)
	if err != nil { t.Fatalf("export --output failed: %v", err) }
	if !strings.Contains(out, "✓ 配置已导出") { t.Errorf("unexpected output: %q", out) }
	data, _ := os.ReadFile(tmpFile)
	if !strings.Contains(string(data), "s1") { t.Errorf("export file missing data") }
}

func TestImportCommand(t *testing.T) {
	mgr := testManager(t)
	importData := `{"connections":[{"name":"imported","host":"9.9.9.9","port":22,"user":"root","password":"pass"}]}`
	tmpFile := filepath.Join(t.TempDir(), "import.json")
	os.WriteFile(tmpFile, []byte(importData), 0600)
	out, err := run(t, mgr, "import", tmpFile)
	if err != nil { t.Fatalf("import failed: %v", err) }
	if !strings.Contains(out, "✓ 成功导入 1 个连接") { t.Errorf("unexpected output: %q", out) }
	showOut, _ := run(t, mgr, "show", "imported")
	if !strings.Contains(showOut, "9.9.9.9") { t.Errorf("import did not work: %q", showOut) }
}

func TestPasswordCommand(t *testing.T) {
	mgr := testManager(t)
	run(t, mgr, "add", "s1", "--host", "1.1.1.1", "--user", "root", "--password", "mysecret")
	out, err := run(t, mgr, "password", "s1")
	if err != nil { t.Fatalf("password failed: %v", err) }
	if !strings.Contains(out, "mysecret") { t.Errorf("password output incorrect: %q", out) }
}

func TestPasswordNonExistent(t *testing.T) {
	mgr := testManager(t)
	_, err := run(t, mgr, "password", "nope")
	if err == nil { t.Error("password for non-existent should fail") }
}

func TestAddShortFlags(t *testing.T) {
	mgr := testManager(t)
	out, err := run(t, mgr, "add", "s1", "-H", "1.1.1.1", "-u", "admin", "-p", "pass", "-t", "prod,web", "-n", "notes")
	if err != nil { t.Fatalf("add with short flags failed: %v", err) }
	if !strings.Contains(out, "✓ 连接 s1 已添加") { t.Errorf("unexpected output: %q", out) }
	// Verify the values were set correctly
	showOut, _ := run(t, mgr, "show", "s1")
	if !strings.Contains(showOut, "admin") { t.Errorf("short flag -u not working: %q", showOut) }
	if !strings.Contains(showOut, "prod") { t.Errorf("short flag -t not working: %q", showOut) }
}

func TestTagsShowConnections(t *testing.T) {
	mgr := testManager(t)
	run(t, mgr, "add", "s1", "--host", "1.1.1.1", "--user", "root", "--password", "pass", "--tag", "prod")
	run(t, mgr, "add", "s2", "--host", "2.2.2.2", "--user", "root", "--password", "pass", "--tag", "prod")
	run(t, mgr, "add", "s3", "--host", "3.3.3.3", "--user", "root", "--password", "pass", "--tag", "dev")

	// Show specific tag
	out, err := run(t, mgr, "tags", "prod")
	if err != nil { t.Fatalf("tags prod failed: %v", err) }
	if !strings.Contains(out, "s1") { t.Errorf("tags prod missing s1: %q", out) }
	if !strings.Contains(out, "s2") { t.Errorf("tags prod missing s2: %q", out) }
	if strings.Contains(out, "s3") { t.Errorf("tags prod should not contain s3: %q", out) }
	if !strings.Contains(out, "1.1.1.1") { t.Errorf("tags prod missing host: %q", out) }
}
