package file

import (
	"fmt"
	"strings"
	"time"

	cmdutil "github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db/file"
	"github.com/GGP1/kure/pb"
	"github.com/GGP1/kure/tree"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

const (
	_ = 1 << (10 * iota)
	// KB - 1024 bytes
	KB
	// MB - 1048576 bytes
	MB
	// GB - 1073741824 bytes
	GB
	// TB - 1099511627776 bytes
	TB
)

var filter bool

var lsExample = `
* List a file and copy its content to the clipboard
kure file ls fileName -c

* Filter files by name
kure file ls fileName -f

* List all files
kure file ls`

func lsSubCmd(db *bolt.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls <name>",
		Short:   "List files",
		Example: lsExample,
		PreRunE: cmdutil.RequirePassword(db),
		RunE:    runLs(db),
		PostRun: func(cmd *cobra.Command, args []string) {
			// Reset flags (session)
			filter = false
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&filter, "filter", "f", false, "filter files by name")

	return cmd
}

func runLs(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")

		switch name {
		case "":
			files, err := file.ListNames(db)
			if err != nil {
				return err
			}
			tree.Print(files)

		default:
			if filter {
				files, err := file.ListNames(db)
				if err != nil {
					return err
				}

				var list []string
				for _, file := range files {
					if strings.Contains(file, name) {
						list = append(list, file)
					}
				}

				if len(list) == 0 {
					return errors.New("no files were found")
				}

				tree.Print(list)
				break
			}

			file, err := file.Get(db, name)
			if err != nil {
				return err
			}

			printFile(file)
		}
		return nil
	}
}

func printFile(f *pb.File) {
	t := time.Unix(f.CreatedAt, 0)
	bytes := len(f.Content)
	size := fmt.Sprintf("%d bytes", bytes)

	switch {
	case bytes >= TB:
		size = fmt.Sprintf("%d TB", bytes/TB)
	case bytes >= GB:
		size = fmt.Sprintf("%d GB", bytes/GB)
	case bytes >= MB:
		size = fmt.Sprintf("%d MB", bytes/MB)
	case bytes >= KB:
		size = fmt.Sprintf("%d KB", bytes/KB)
	}

	fields := map[string]string{
		"Path":       f.Name,
		"Filename":   f.Filename,
		"Size":       size,
		"Created at": fmt.Sprintf("%v", t),
	}

	box := cmdutil.BuildBox(f.Name, fields)
	fmt.Println("\n" + box)
}
