package main

import (
	"errors"
	stdlog "log"
	"os"

	"github.com/meltwater/drone-cache/cache"
	"github.com/meltwater/drone-cache/cache/backend"
	"github.com/meltwater/drone-cache/internal"
	"github.com/meltwater/drone-cache/metadata"
	"github.com/meltwater/drone-cache/plugin"

	"github.com/go-kit/kit/log/level"
	"github.com/urfave/cli"
)

var version = "0.0.0"

//nolint:funlen
func main() {
	app := cli.NewApp()
	app.Name = "Drone cache plugin"
	app.Usage = "Drone cache plugin"
	app.Action = run
	app.Version = version
	app.Flags = []cli.Flag{
		// Logger args

		cli.StringFlag{
			Name:   "log.level, ll",
			Usage:  "log filtering level. ('error', 'warn', 'info', 'debug')",
			Value:  internal.LogLevelInfo,
			EnvVar: "PLUGIN_LOG_LEVEL, LOG_LEVEL",
		},
		cli.StringFlag{
			Name:   "log.format, lf",
			Usage:  "log format to use. ('logfmt', 'json')",
			Value:  internal.LogFormatLogfmt,
			EnvVar: "PLUGIN_LOG_FORMAT, LOG_FORMAT",
		},

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
			Name:   "repo.owner, ro",
			Usage:  "repository owner (for Drone version < 1.0)",
			EnvVar: "DRONE_REPO_OWNER",
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
			Value:  cache.DefaultArchiveFormat,
			EnvVar: "PLUGIN_ARCHIVE_FORMAT",
		},
		cli.IntFlag{
			Name: "compression-level, cpl",
			Usage: `compression level to use for gzip compression when archive-format specified as gzip
			(check https://godoc.org/compress/flate#pkg-constants for available options)`,
			Value:  cache.DefaultCompressionLevel,
			EnvVar: "PLUGIN_COMPRESSION_LEVEL",
		},
		cli.BoolFlag{
			Name:   "skip-symlinks, ss",
			Usage:  "skip symbolic links in archive",
			EnvVar: "PLUGIN_SKIP_SYMLINKS, SKIP_SYMLINKS",
		},
		cli.BoolFlag{
			Name:   "debug, d",
			Usage:  "debug",
			EnvVar: "PLUGIN_DEBUG, DEBUG",
		},
		cli.BoolFlag{
			Name:   "exit-code, ex",
			Usage:  "always exit with exit code, disable silent fails for known errors",
			Hidden: true,
			EnvVar: "PLUGIN_EXIT_CODE, EXIT_CODE",
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
			Usage:  "endpoint for the s3/cloud storage connection",
			EnvVar: "PLUGIN_ENDPOINT,S3_ENDPOINT,CLOUD_STORAGE_ENDPOINT",
		},
		cli.StringFlag{
			Name:   "access-key, akey",
			Usage:  "AWS access key",
			EnvVar: "PLUGIN_ACCESS_KEY,AWS_ACCESS_KEY_ID,CACHE_AWS_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name:   "secret-key, skey",
			Usage:  "AWS/GCP secret key",
			EnvVar: "PLUGIN_SECRET_KEY,AWS_SECRET_ACCESS_KEY,CACHE_AWS_SECRET_ACCESS_KEY,GCP_API_KEY",
		},
		cli.StringFlag{
			Name:   "bucket, bckt",
			Usage:  "AWS bucket name",
			EnvVar: "PLUGIN_BUCKET,S3_BUCKET,CLOUD_STORAGE_BUCKET",
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

		// Azure specific Config flags

		cli.StringFlag{
			Name:   "azure-account-name",
			Usage:  "Azure Blob Storage Account Name",
			EnvVar: "PLUGIN_ACCOUNT_NAME,AZURE_ACCOUNT_NAME",
		},
		cli.StringFlag{
			Name:   "azure-account-key",
			Usage:  "Azure Blob Storage Account Key",
			EnvVar: "PLUGIN_ACCOUNT_KEY,AZURE_ACCOUNT_KEY",
		},
		cli.StringFlag{
			Name:   "azure-container-name",
			Usage:  "Azure Blob Storage container name",
			EnvVar: "PLUGIN_CONTAINER,AZURE_CONTAINER_NAME",
		},
		cli.StringFlag{
			Name:   "azure-blob-storage-url",
			Usage:  "Azure Blob Storage URL",
			Value:  "blob.core.windows.net",
			EnvVar: "AZURE_BLOB_STORAGE_URL",
		},

		// SFTP specific Config flags

		cli.StringFlag{
			Name:   "sftp-cache-root",
			Usage:  "sftp root directory",
			EnvVar: "SFTP_CACHE_ROOT",
		},
		cli.StringFlag{
			Name:   "sftp-username",
			Usage:  "sftp username",
			EnvVar: "SFTP_USERNAME",
		},
		cli.StringFlag{
			Name:   "sftp-password",
			Usage:  "sftp password",
			EnvVar: "SFTP_PASSWORD",
		},
		cli.StringFlag{
			Name:   "ftp-public-key-file",
			Usage:  "sftp public key file path",
			EnvVar: "SFTP_PUBLIC_KEY_FILE",
		},
		cli.StringFlag{
			Name:   "sftp-auth-method",
			Usage:  "sftp auth method, defaults to none. (PASSWORD, PUBLIC_KEY_FILE)",
			EnvVar: "SFTP_AUTH_METHOD",
		},
		cli.StringFlag{
			Name:   "sftp-host",
			Usage:  "sftp host",
			EnvVar: "SFTP_HOST",
		},
		cli.StringFlag{
			Name:   "sftp-port",
			Usage:  "sftp port",
			EnvVar: "SFTP_PORT",
		},
	}

	if err := app.Run(os.Args); err != nil {
		stdlog.Fatalf("%+v", err)
	}
}

//nolint:funlen
func run(c *cli.Context) error {
	var logLevel = c.String("log.level")
	if c.Bool("debug") {
		logLevel = internal.LogLevelDebug
	}

	logger := internal.NewLogger(logLevel, c.String("log.format"), "drone-cache")

	plg := plugin.Plugin{
		Logger: logger,
		Metadata: metadata.Metadata{
			Repo: metadata.Repo{
				Namespace: c.String("repo.namespace"),
				Owner:     c.String("repo.owner"),
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
			ArchiveFormat:    c.String("archive-format"),
			Backend:          c.String("backend"),
			CacheKey:         c.String("cache-key"),
			CompressionLevel: c.Int("compression-level"),
			Debug:            c.Bool("debug"),
			Mount:            c.StringSlice("mount"),
			Rebuild:          c.Bool("rebuild"),
			Restore:          c.Bool("restore"),

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

			Azure: backend.AzureConfig{
				AccountName:    c.String("azure-account-name"),
				AccountKey:     c.String("azure-account-key"),
				ContainerName:  c.String("azure-container-name"),
				BlobStorageURL: c.String("azure-blob-storage-url"),
				Azurite:        false,
			},

			SFTP: backend.SFTPConfig{
				CacheRoot: c.String("sftp-cache-root"),
				Username:  c.String("sftp-username"),
				Host:      c.String("sftp-host"),
				Port:      c.String("sftp-port"),
				Auth: backend.SSHAuth{
					Password:      c.String("sftp-password"),
					PublicKeyFile: c.String("sftp-public-key-file"),
					Method:        backend.SSHAuthMethod(c.String("sftp-auth-method")),
				},
			},

			CloudStorage: backend.CloudStorageConfig{
				Bucket:     c.String("bucket"),
				Encryption: c.String("encryption"),
				Endpoint:   c.String("endpoint"),
				APIKey:     c.String("secret-key"),
			},

			SkipSymlinks: c.Bool("skip-symlinks"),
		},
	}

	err := plg.Exec()
	if err == nil {
		return nil
	}

	if c.Bool("exit-code") {
		// If it is exit-code enabled, always exit with error.
		level.Warn(logger).Log("msg", "silent fails disabled, exiting with status code on error")
		return err
	}

	var e plugin.Error
	if errors.As(err, &e) {
		// If it is an expected error log it, handle it gracefully,
		level.Error(logger).Log("err", err)

		return nil
	}

	return err
}
