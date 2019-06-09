package main

import (
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/meltwater/drone-cache/cache/backend"
	"github.com/meltwater/drone-cache/metadata"
	"github.com/meltwater/drone-cache/plugin"
)

func main() {
	app := cli.NewApp()
	app.Name = "Drone cache plugin"
	app.Usage = "Drone cache plugin"
	app.Action = run
	app.Version = "1.0.3"
	app.Flags = []cli.Flag{
		// Repo args

		cli.StringFlag{
			Name:   "repo.fullname, rf",
			Usage:  "repository full name",
			EnvVar: "DRONE_REPO",
		},
		cli.StringFlag{
			Name:   "repo.namespace, rns",
			Usage:  "repository namespace",
			EnvVar: "DRONE_REPO_NAMESPACE",
		},
		cli.StringFlag{
			Name:   "repo.name, rn",
			Usage:  "repository name",
			EnvVar: "DRONE_REPO_NAME",
		},
		cli.StringFlag{
			Name:   "repo.link, rl",
			Usage:  "repository link",
			EnvVar: "DRONE_REPO_LINK",
		},
		cli.StringFlag{
			Name:   "repo.avatar, ra",
			Usage:  "repository avatar",
			EnvVar: "DRONE_REPO_AVATAR",
		},
		cli.StringFlag{
			Name:   "repo.branch, rb",
			Usage:  "repository default branch",
			EnvVar: "DRONE_REPO_BRANCH",
		},
		cli.BoolFlag{
			Name:   "repo.private, rp",
			Usage:  "repository is private",
			EnvVar: "DRONE_REPO_PRIVATE",
		},
		cli.BoolFlag{
			Name:   "repo.trusted, rt",
			Usage:  "repository is trusted",
			EnvVar: "DRONE_REPO_TRUSTED",
		},

		// Commit args

		cli.StringFlag{
			Name:   "remote.url, remu",
			Usage:  "git remote url",
			EnvVar: "DRONE_REMOTE_URL",
		},
		cli.StringFlag{
			Name:   "commit.sha, cs",
			Usage:  "git commit sha",
			EnvVar: "DRONE_COMMIT_SHA",
		},
		cli.StringFlag{
			Name:   "commit.ref, cr",
			Value:  "refs/heads/master",
			Usage:  "git commit ref",
			EnvVar: "DRONE_COMMIT_REF",
		},
		cli.StringFlag{
			Name:   "commit.branch, cb",
			Value:  "master",
			Usage:  "git commit branch",
			EnvVar: "DRONE_COMMIT_BRANCH",
		},
		cli.StringFlag{
			Name:   "commit.message, cm",
			Usage:  "git commit message",
			EnvVar: "DRONE_COMMIT_MESSAGE",
		},
		cli.StringFlag{
			Name:   "commit.link, cl",
			Usage:  "git commit link",
			EnvVar: "DRONE_COMMIT_LINK",
		},
		cli.StringFlag{
			Name:   "commit.author.name, an",
			Usage:  "git author name",
			EnvVar: "DRONE_COMMIT_AUTHOR",
		},
		cli.StringFlag{
			Name:   "commit.author.email, ae",
			Usage:  "git author email",
			EnvVar: "DRONE_COMMIT_AUTHOR_EMAIL",
		},
		cli.StringFlag{
			Name:   "commit.author.avatar, aa",
			Usage:  "git author avatar",
			EnvVar: "DRONE_COMMIT_AUTHOR_AVATAR",
		},

		// Build args

		cli.StringFlag{
			Name:   "build.event, be",
			Value:  "push",
			Usage:  "build event",
			EnvVar: "DRONE_BUILD_EVENT",
		},
		cli.IntFlag{
			Name:   "build.number, bn",
			Usage:  "build number",
			EnvVar: "DRONE_BUILD_NUMBER",
		},
		cli.IntFlag{
			Name:   "build.created, bc",
			Usage:  "build created",
			EnvVar: "DRONE_BUILD_CREATED",
		},
		cli.IntFlag{
			Name:   "build.started, bs",
			Usage:  "build started",
			EnvVar: "DRONE_BUILD_STARTED",
		},
		cli.IntFlag{
			Name:   "build.finished, bf",
			Usage:  "build finished",
			EnvVar: "DRONE_BUILD_FINISHED",
		},
		cli.StringFlag{
			Name:   "build.status, bstat",
			Usage:  "build status",
			Value:  "success",
			EnvVar: "DRONE_BUILD_STATUS",
		},
		cli.StringFlag{
			Name:   "build.link, bl",
			Usage:  "build link",
			EnvVar: "DRONE_BUILD_LINK",
		},
		cli.StringFlag{
			Name:   "build.deploy, db",
			Usage:  "build deployment target",
			EnvVar: "DRONE_DEPLOY_TO",
		},
		cli.BoolFlag{
			Name:   "yaml.verified, yv",
			Usage:  "build yaml is verified",
			EnvVar: "DRONE_YAML_VERIFIED",
		},
		cli.BoolFlag{
			Name:   "yaml.signed, ys",
			Usage:  "build yaml is signed",
			EnvVar: "DRONE_YAML_SIGNED",
		},

		// Prev build args

		cli.IntFlag{
			Name:   "prev.build.number, pbn",
			Usage:  "previous build number",
			EnvVar: "DRONE_PREV_BUILD_NUMBER",
		},
		cli.StringFlag{
			Name:   "prev.build.status, pbst",
			Usage:  "previous build status",
			EnvVar: "DRONE_PREV_BUILD_STATUS",
		},
		cli.StringFlag{
			Name:   "prev.commit.sha, pcs",
			Usage:  "previous build sha",
			EnvVar: "DRONE_PREV_COMMIT_SHA",
		},

		// Config args

		cli.StringFlag{
			Name:   "backend, b",
			Usage:  "cache backend to use in plugin (s3, filesystem)",
			Value:  "s3",
			EnvVar: "PLUGIN_BACKEND",
		},

		cli.StringSliceFlag{
			Name:   "mount, m",
			Usage:  "cache directories, an array of folders to cache",
			EnvVar: "PLUGIN_MOUNT",
		},
		cli.BoolFlag{
			Name:   "rebuild, reb",
			Usage:  "rebuild the cache directories",
			EnvVar: "PLUGIN_REBUILD",
		},
		cli.BoolFlag{
			Name:   "restore, res",
			Usage:  "restore the cache directories",
			EnvVar: "PLUGIN_RESTORE",
		},
		cli.StringFlag{
			Name:   "cache-key, chk",
			Usage:  "cache key to use for the cache directories",
			EnvVar: "PLUGIN_CACHE_KEY",
		},
		cli.StringFlag{
			Name:   "archive-format, arcfmt",
			Usage:  "archive format to use to store the cache directories (tar, gzip)",
			Value:  "tar",
			EnvVar: "PLUGIN_ARCHIVE_FORMAT",
		},
		cli.StringFlag{
			Name:   "debug, d",
			Usage:  "debug",
			EnvVar: "PLUGIN_DEBUG, DEBUG",
		},

		// Volume specific Config args

		cli.StringFlag{
			Name:   "filesystem-cache-root, fcr",
			Usage:  "local filesystem root directory for the filesystem cache",
			Value:  "/tmp/cache",
			EnvVar: "PLUGIN_FILESYSTEM_CACHE_ROOT, FILESYSTEM_CACHE_ROOT",
		},

		// S3 specific Config args

		cli.StringFlag{
			Name:   "endpoint, e",
			Usage:  "endpoint for the s3 connection",
			EnvVar: "PLUGIN_ENDPOINT,S3_ENDPOINT",
		},
		cli.StringFlag{
			Name:   "access-key, akey",
			Usage:  "AWS access key",
			EnvVar: "PLUGIN_ACCESS_KEY,AWS_ACCESS_KEY_ID,CACHE_AWS_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name:   "secret-key, skey",
			Usage:  "AWS secret key",
			EnvVar: "PLUGIN_SECRET_KEY,AWS_SECRET_ACCESS_KEY,CACHE_AWS_SECRET_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "bucket, bckt",
			Usage:  "AWS bucket name",
			EnvVar: "PLUGIN_BUCKET,S3_BUCKET",
		},
		cli.StringFlag{
			Name:   "region, reg",
			Usage:  "AWS bucket region. (us-east-1, eu-west-1, ...)",
			EnvVar: "PLUGIN_REGION,S3_REGION",
		},
		cli.BoolFlag{
			Name:   "path-style, ps",
			Usage:  "use path style for bucket paths. (true for minio, false for aws)",
			EnvVar: "PLUGIN_PATH_STYLE",
		},
		cli.StringFlag{
			Name:   "acl",
			Usage:  "upload files with acl (private, public-read, ...)",
			Value:  "private",
			EnvVar: "PLUGIN_ACL",
		},
		cli.StringFlag{
			Name:   "encryption, enc",
			Usage:  "server-side encryption algorithm, defaults to none. (AES256, aws:kms)",
			EnvVar: "PLUGIN_ENCRYPTION",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("%+v", err)
	}
}

func run(c *cli.Context) error {
	plg := plugin.Plugin{
		Metadata: metadata.Metadata{
			Repo: metadata.Repo{
				Namespace: c.String("repo.namespace"),
				Owner:     c.String("repo.namespace"),
				Name:      c.String("repo.name"),
				Link:      c.String("repo.link"),
				Avatar:    c.String("repo.avatar"),
				Branch:    c.String("repo.branch"),
				Private:   c.Bool("repo.private"),
				Trusted:   c.Bool("repo.trusted"),
			},
			Build: metadata.Build{
				Number:   c.Int("build.number"),
				Event:    c.String("build.event"),
				Status:   c.String("build.status"),
				Deploy:   c.String("build.deploy"),
				Created:  int64(c.Int("build.created")),
				Started:  int64(c.Int("build.started")),
				Finished: int64(c.Int("build.finished")),
				Link:     c.String("build.link"),
			},
			Commit: metadata.Commit{
				Remote:  c.String("remote.url"),
				Sha:     c.String("commit.sha"),
				Ref:     c.String("commit.sha"),
				Link:    c.String("commit.link"),
				Branch:  c.String("commit.branch"),
				Message: c.String("commit.message"),
				Author: metadata.Author{
					Name:   c.String("commit.author.name"),
					Email:  c.String("commit.author.email"),
					Avatar: c.String("commit.author.avatar"),
				},
			},
		},
		Config: plugin.Config{
			ArchiveFormat: c.String("archive-format"),
			Backend:       c.String("backend"),
			CacheKey:      c.String("cache-key"),
			Debug:         c.Bool("debug"),
			Mount:         c.StringSlice("mount"),
			Rebuild:       c.Bool("rebuild"),
			Restore:       c.Bool("restore"),

			FileSystem: backend.FileSystemConfig{
				CacheRoot: c.String("filesystem-cache-root"),
			},
			S3: backend.S3Config{
				ACL:        c.String("acl"),
				Bucket:     c.String("bucket"),
				Encryption: c.String("encryption"),
				Endpoint:   c.String("endpoint"),
				Key:        c.String("access-key"),
				PathStyle:  c.Bool("path-style"),
				Region:     c.String("region"),
				Secret:     c.String("secret-key"),
			},
		},
	}

	err := plg.Exec()
	if _, ok := err.(plugin.Error); ok {
		// If it is an expected error log it, handle it gracefully
		log.Println(err)
		return nil
	}
	return err
}
