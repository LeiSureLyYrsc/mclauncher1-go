package launch

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/LeiSureLyYrsc/mclauncher1-go/internal/auth"
	"github.com/LeiSureLyYrsc/mclauncher1-go/internal/bmclapi"
)

// Options for assembling a launch command.
type Options struct {
	GameDir     string
	VersionID   string
	JavaPath    string
	MaxMemoryMB int
	Session     auth.Session
	AuthlibJar  string // path to authlib-injector.jar (optional)
	Debug       bool
	Version     *bmclapi.VersionJSON
}

// Command is the final java invocation.
type Command struct {
	Java string
	Args []string
	Dir  string
}

// Build constructs JVM + game arguments for the given version metadata.
func Build(opt Options) (Command, error) {
	vj := opt.Version
	if vj == nil {
		return Command{}, fmt.Errorf("version metadata missing")
	}
	id := opt.VersionID
	if id == "" {
		id = vj.ID
	}
	versionsDir := filepath.Join(opt.GameDir, "versions", id)
	jarPath := filepath.Join(versionsDir, id+".jar")
	nativesDir := filepath.Join(versionsDir, "natives")
	assetsDir := filepath.Join(opt.GameDir, "assets")
	libsDir := filepath.Join(opt.GameDir, "libraries")

	cp, err := classpath(vj, libsDir, jarPath)
	if err != nil {
		return Command{}, err
	}

	mem := opt.MaxMemoryMB
	if mem <= 0 {
		mem = 2048
	}
	java := opt.JavaPath
	if java == "" {
		java = "java"
	}

	vars := map[string]string{
		"auth_player_name":  opt.Session.DisplayName(),
		"version_name":      id,
		"game_directory":    opt.GameDir,
		"assets_root":       assetsDir,
		"assets_index_name": assetsIndexName(vj),
		"auth_uuid":         strings.ReplaceAll(opt.Session.UUID, "-", ""),
		"auth_access_token": opt.Session.AccessToken,
		"user_type":         opt.Session.UserType,
		"version_type":      vj.Type,
		"natives_directory": nativesDir,
		"launcher_name":     "mclauncher1-go",
		"launcher_version":  "0.1",
		"classpath":         cp,
		"library_directory": libsDir,
		"classpath_separator": classpathSep(),
	}
	if vars["user_type"] == "" {
		vars["user_type"] = "legacy"
	}
	if vars["auth_access_token"] == "" {
		vars["auth_access_token"] = "0"
	}

	var jvmArgs []string
	jvmArgs = append(jvmArgs, fmt.Sprintf("-Xmx%dM", mem))
	jvmArgs = append(jvmArgs, fmt.Sprintf("-Djava.library.path=%s", nativesDir))
	if opt.Session.Mode == auth.ModeAuthlib && opt.AuthlibJar != "" && opt.Session.AuthlibAPI != "" {
		jvmArgs = append(jvmArgs, fmt.Sprintf("-javaagent:%s=%s", opt.AuthlibJar, opt.Session.AuthlibAPI))
	}

	if vj.Arguments != nil && len(vj.Arguments.JVM) > 0 {
		jvmArgs = append(jvmArgs, expandArgList(vj.Arguments.JVM, vars)...)
	} else {
		// legacy defaults
		if runtime.GOOS == "darwin" {
			jvmArgs = append(jvmArgs, "-XstartOnFirstThread")
		}
		jvmArgs = append(jvmArgs, "-cp", cp)
	}

	// Ensure -cp present
	hasCP := false
	for i, a := range jvmArgs {
		if a == "-cp" || a == "-classpath" {
			hasCP = true
			if i+1 < len(jvmArgs) && (jvmArgs[i+1] == "" || strings.Contains(jvmArgs[i+1], "${")) {
				jvmArgs[i+1] = cp
			}
			break
		}
	}
	if !hasCP {
		jvmArgs = append(jvmArgs, "-cp", cp)
	}

	mainClass := vj.MainClass
	if mainClass == "" {
		return Command{}, fmt.Errorf("mainClass empty")
	}

	var gameArgs []string
	if vj.Arguments != nil && len(vj.Arguments.Game) > 0 {
		gameArgs = expandArgList(vj.Arguments.Game, vars)
	} else if vj.MinecraftArguments != "" {
		gameArgs = expandTokens(strings.Fields(vj.MinecraftArguments), vars)
	} else {
		gameArgs = []string{
			"--username", vars["auth_player_name"],
			"--version", vars["version_name"],
			"--gameDir", vars["game_directory"],
			"--assetsDir", vars["assets_root"],
			"--assetIndex", vars["assets_index_name"],
			"--uuid", vars["auth_uuid"],
			"--accessToken", vars["auth_access_token"],
			"--userType", vars["user_type"],
			"--versionType", vars["version_type"],
		}
	}

	args := append(jvmArgs, mainClass)
	args = append(args, gameArgs...)
	return Command{Java: java, Args: args, Dir: opt.GameDir}, nil
}

func assetsIndexName(vj *bmclapi.VersionJSON) string {
	if vj.AssetIndex != nil && vj.AssetIndex.ID != "" {
		return vj.AssetIndex.ID
	}
	if vj.Assets != "" {
		return vj.Assets
	}
	return "legacy"
}

func classpath(vj *bmclapi.VersionJSON, libsDir, clientJar string) (string, error) {
	var parts []string
	for _, lib := range vj.Libraries {
		if !bmclapi.LibraryAllowed(lib.Rules) {
			continue
		}
		// skip pure natives jars from classpath (artifact may still exist)
		path := ""
		if lib.Downloads != nil && lib.Downloads.Artifact != nil {
			path = lib.Downloads.Artifact.Path
		}
		if path == "" && lib.Name != "" {
			path = mavenPath(lib.Name)
		}
		if path == "" {
			continue
		}
		parts = append(parts, filepath.Join(libsDir, filepath.FromSlash(path)))
	}
	parts = append(parts, clientJar)
	return strings.Join(parts, classpathSep()), nil
}

func classpathSep() string {
	if runtime.GOOS == "windows" {
		return ";"
	}
	return ":"
}

func mavenPath(name string) string {
	parts := strings.Split(name, ":")
	if len(parts) < 3 {
		return name
	}
	group := strings.ReplaceAll(parts[0], ".", "/")
	artifact, version := parts[1], parts[2]
	file := artifact + "-" + version + ".jar"
	if len(parts) >= 4 {
		file = artifact + "-" + version + "-" + parts[3] + ".jar"
	}
	return group + "/" + artifact + "/" + version + "/" + file
}

func expandArgList(raw []any, vars map[string]string) []string {
	var out []string
	for _, item := range raw {
		switch v := item.(type) {
		case string:
			out = append(out, expandOne(v, vars))
		case map[string]any:
			if !argAllowed(v) {
				continue
			}
			val := v["value"]
			switch tv := val.(type) {
			case string:
				out = append(out, expandOne(tv, vars))
			case []any:
				for _, x := range tv {
					if s, ok := x.(string); ok {
						out = append(out, expandOne(s, vars))
					}
				}
			}
		default:
			// json.Number etc — re-marshal
			b, _ := json.Marshal(item)
			var m map[string]any
			if json.Unmarshal(b, &m) == nil {
				if !argAllowed(m) {
					continue
				}
				val := m["value"]
				switch tv := val.(type) {
				case string:
					out = append(out, expandOne(tv, vars))
				case []any:
					for _, x := range tv {
						if s, ok := x.(string); ok {
							out = append(out, expandOne(s, vars))
						}
					}
				}
			}
		}
	}
	return out
}

func argAllowed(m map[string]any) bool {
	rulesRaw, ok := m["rules"]
	if !ok {
		return true
	}
	b, _ := json.Marshal(rulesRaw)
	var rules []bmclapi.Rule
	if json.Unmarshal(b, &rules) != nil {
		return true
	}
	return bmclapi.LibraryAllowed(rules)
}

func expandTokens(in []string, vars map[string]string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = expandOne(s, vars)
	}
	return out
}

func expandOne(s string, vars map[string]string) string {
	for k, v := range vars {
		s = strings.ReplaceAll(s, "${"+k+"}", v)
	}
	return s
}
