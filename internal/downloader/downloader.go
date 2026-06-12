package downloader

import (
	"context"
	// "sync"
	"errors"
	"fmt"
)

type Downloader[T any] struct {
	filesChan chan *T
	workers   []*Worker[T]
}

func New[T any](ctx Context.context, maxNumFiles int, workersNum int) {
	workers := make([]*Worker[T], workersNum)
	filesChan := make(chan *T, maxNumFiles)

	// create workers
	for i := range workersNum {
		var w = createWorker[T](i, filesChan)
		go w.run(ctx)
		workers[i] = w
	}
	return &Downloader{
		filesChan: filesChan,
		workers: workers,
	}
}

// sum: md5sum
// if check sum provided: check sum of a file in filePath -> skip if already exists
func downloadFile(ctx context.Context, filePath string, url string, sum *string) error {
	if sum == nil {
		fmt.Printf("RISKY! No sum provided -> unable to check for a file\n")
	} else {
		_, err := os.Stat(filePath)
		if !os.IsNotExist(err) {
			exists, err := checkSum(filePath, sum)
			if err != nil {
				fmt.Printf("Unable to check sum for file [%s], then downloading it ...\n", filePath)
			}
			if exists {
				fmt.Printf("File [%s] exists! Skipping...\n", filePath)
				return nil
			}
		}
	}

	out, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Unable to create file [%s]\n", filePath)
		return err
	}
	defer out.Close()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		fmt.Printf("Unable to make req [%s]\n", err)
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Unable to get url [%s]\n", url)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Errorf("bad status: %s", resp.Status)
		return err
	}
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("Unable to write into file [%s]\n", filePath)
		return err
	}

	// check sum post download
	_, err = checkSum(filePath, sum)
	if err != nil {
		fmt.Printf("Smth went wrong. Check sum failed\n")
		return err
	}
	return nil
}

// comapre file by filePath with md5 string
func checkSum(filePath string, sum *string) (bool, error) {
	fileSum, err := md5Sum(filePath)
	if err != nil {
		fmt.Printf("Unable to calcaulte md5sum for file [%s]\n", filePath)
		return false, err
	}
	if fileSum == *sum {
		return true, nil
	}
	return false, nil
}

func md5Sum(filePath string) (string, error) {
	fi, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Unable to open file [%s]\n", filePath)
		return "", err
	}
	defer fi.Close()

	h := md5.New()
	for {
		var data [256]byte
		n, err := fi.Read(data[:])
		if n == 0 && err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("ERROR [%s]: %s", filePath, err)
			return "", err
		}
		h.Write(data[0:n])
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

