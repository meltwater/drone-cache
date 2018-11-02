package main

import (
	"log"
	"os"

	"github.com/meltwater/drone-s3-cache/plugin"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "S3 cache plugin"
	app.Usage = "S3 cache plugin"
	app.Action = run
	app.Version = "0.9.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "repo.name",
			Usage:  "repository full name",
			EnvVar: "DRONE_REPO",
		},
		cli.StringFlag{
			Name:   "repo.branch",
			Usage:  "repository default branch",
			EnvVar: "DRONE_REPO_BRANCH",
		},
		cli.StringFlag{
			Name:   "commit.branch",
			Value:  "master",
			Usage:  "repository branch",
			EnvVar: "DRONE_COMMIT_BRANCH",
		},
		cli.StringSliceFlag{
			Name:   "mount",
			Usage:  "cache directories",
			EnvVar: "PLUGIN_MOUNT",
		},
		cli.BoolFlag{
			Name:   "rebuild",
			Usage:  "rebuild the cache directories",
			EnvVar: "PLUGIN_REBUILD",
		},
		cli.BoolFlag{
			Name:   "restore",
			Usage:  "restore the cache directories",
			EnvVar: "PLUGIN_RESTORE",
		},
		cli.StringFlag{
			Name:   "endpoint",
			Usage:  "endpoint for the s3 connection",
			EnvVar: "PLUGIN_ENDPOINT,S3_ENDPOINT",
		},
		cli.StringFlag{
			Name:   "access-key",
			Usage:  "aws access key",
			EnvVar: "PLUGIN_ACCESS_KEY,AWS_ACCESS_KEY_ID,CACHE_AWS_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name:   "secret-key",
			Usage:  "aws secret key",
			EnvVar: "PLUGIN_SECRET_KEY,AWS_SECRET_ACCESS_KEY,CACHE_AWS_SECRET_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "bucket",
			Usage:  "aws bucket",
			Value:  "mw-drone-prod-ew1dro1",
			EnvVar: "PLUGIN_BUCKET,S3_BUCKET",
		},
		cli.StringFlag{
			Name:   "region",
			Usage:  "aws region",
			Value:  "us-east-1",
			EnvVar: "PLUGIN_REGION,S3_REGION",
		},
		cli.BoolFlag{
			Name:   "path-style",
			Usage:  "use path style for bucket paths",
			EnvVar: "PLUGIN_PATH_STYLE",
		},
		cli.StringFlag{
			Name:   "acl",
			Usage:  "upload files with acl",
			Value:  "private",
			EnvVar: "PLUGIN_ACL",
		},
		cli.StringFlag{
			Name:   "encryption",
			Usage:  "server-side encryption algorithm, defaults to none",
			EnvVar: "PLUGIN_ENCRYPTION",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("%+v", err)
	}
}

func run(c *cli.Context) error {
	plugin := plugin.Plugin{
		ACL:        c.String("acl"),
		Branch:     c.String("commit.branch"),
		Bucket:     c.String("bucket"),
		Default:    c.String("repo.branch"),
		Encryption: c.String("encryption"),
		Endpoint:   c.String("endpoint"),
		Key:        c.String("access-key"),
		Mount:      c.StringSlice("mount"),
		PathStyle:  c.Bool("path-style"),
		Rebuild:    c.Bool("rebuild"),
		Region:     c.String("region"),
		Repo:       c.String("repo.name"),
		Restore:    c.Bool("restore"),
		Secret:     c.String("secret-key"),
	}

	return plugin.Exec()
}
