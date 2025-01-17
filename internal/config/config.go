package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
)

var cfg *koanf.Koanf

const (
	CMD               = "cmd"
	LOG_LEVEL         = "log.level"
	CALDAV_URL        = "caldav.url"
	CALDAV_USER       = "caldav.user"
	CALDAV_PASS       = "caldav.pass"
	TODOIST_TOKEN     = "todoist.token"
	TODOIST_PROJECTID = "todoist.projectid"
	prefix            = "CALSYNC_"
)

func Gist() *koanf.Koanf {
	if cfg == nil {
		ini()
	}
	return cfg
}

func Sprint() string {
	sb := strings.Builder{}
	sb.WriteString("cmd|required|-\n")
	sb.WriteString("log_level|optional|info\n")
	sb.WriteString("caldav_url|required|-\n")
	sb.WriteString("caldav_user|optional|-\n")
	sb.WriteString("caldav_pass|optional|-\n")
	sb.WriteString("todoist_token|required|-\n")
	return sb.String()
}

func ini() {
	cfg = koanf.New(".")
	cfg.Set(LOG_LEVEL, "info")

	f := flag.NewFlagSet("config", flag.ContinueOnError)
	f.Usage = func() {
		fmt.Println(f.FlagUsages())
		os.Exit(0)
	}

	f.String(CMD, "", "application run mode")
	f.String(LOG_LEVEL, "info", "log level")
	f.String(CALDAV_URL, "", "caldav url")
	f.String(CALDAV_USER, "", "caldav user")
	f.String(CALDAV_PASS, "", "caldav password")
	f.String(TODOIST_TOKEN, "", "todoist api token")
	f.String(TODOIST_PROJECTID, "", "todoist project id")
	f.Parse(os.Args[1:])
	if err := cfg.Load(posflag.Provider(f, ".", cfg), nil); err != nil {
		log.Panic().Err(err).Msg("error loading config")
	}
	lvl, err := zerolog.ParseLevel(cfg.String(LOG_LEVEL))
	if err != nil {
		log.Panic().Err(err).Msg("error parsing log level")
	}
	zerolog.SetGlobalLevel(lvl)

	printCfg()
	// Load environment variables
	// cfg.Load(env.Provider(prefix, ".", func(s string) string {
	// 	return strings.Replace(strings.ToLower(
	// 		strings.TrimPrefix(s, prefix)), "_", ".", -1)
	// }), nil)
}

func printCfg() {
	log.Debug().Msgf("cmd: %s", cfg.String(CMD))
	log.Debug().Msgf("log_level: %s", cfg.String(LOG_LEVEL))
	log.Debug().Msgf("caldav_url: %s", cfg.String(CALDAV_URL))
	log.Debug().Msgf("caldav_user: %s", cfg.String(CALDAV_USER))
	log.Debug().Msgf("caldav_pass: %s", cfg.String(CALDAV_PASS))
	log.Debug().Msgf("todoist_token: %s", cfg.String(TODOIST_TOKEN))
	log.Debug().Msgf("todoist_projectid: %s", cfg.String(TODOIST_PROJECTID))
}
