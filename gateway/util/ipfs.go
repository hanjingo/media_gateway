package util

import (
	"context"
	"fmt"
	"os"

	shell "github.com/ipfs/go-ipfs-api"
)

type IpfsUploader struct{}

func NewIpfsUploader() *IpfsUploader {
	return &IpfsUploader{}
}

func (up *IpfsUploader) Upload(ipfsAddr, filename string) (string, error) {
	sh := shell.NewShell(ipfsAddr)
	if sh == nil {
		return "", fmt.Errorf("new shell with addr:%s fail", ipfsAddr)
	}
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return sh.Add(f)
}

type IpfsDownloader struct{}

func NewIpfsDownloader() *IpfsDownloader {
	return &IpfsDownloader{}
}

func (d *IpfsDownloader) Download(ipfsAddr, hash, target string) error {
	sh := shell.NewShell(ipfsAddr)
	if err := sh.Get(hash, target); err != nil {
		return err
	}
	return nil
}

func (d *IpfsDownloader) DownloadAndRmBlock(ipfsAddr, hash, target string) error {
	sh := shell.NewShell(ipfsAddr)
	if err := sh.Get(hash, target); err != nil {
		return err
	}
	if err := sh.Unpin(hash); err != nil {
		return err
	}
	if err := d.BlockRm(ipfsAddr, hash); err != nil {
		return err
	}
	return nil
}

func (s *IpfsDownloader) BlockRm(ipfsAddr, id string) error {
	sh := shell.NewShell(ipfsAddr)
	if sh == nil {
		return fmt.Errorf("new shell with addr:%s fail", ipfsAddr)
	}
	resp, err := sh.Request("block/rm", id).Send(context.Background())
	if err != nil {
		return err
	}
	defer resp.Close()

	if resp.Error != nil {
		return resp.Error
	}
	return nil
}
