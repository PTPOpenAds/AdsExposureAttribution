/*
 * copyright (c) 2020, Tencent Inc.
 * All rights reserved.
 *
 * Author:  linceyou@tencent.com
 * Last Modify: 10/9/20, 5:26 PM
 */

package utility

import (
	"bufio"
	"errors"
	"flag"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
)

var (
	ErrQueueTimeout    = errors.New("queue full")
	DefaultNlines      = flag.Int("default_nlines", 1000, "default_nlines")
	defaultQueueSize   = flag.Int("default_queue_size", 1000, "default_queue_size")
	defaultParallelism = flag.Int("default_parallelism", 1000, "default_parallelism")
	touchSuccess       = flag.Bool("touch", false, "touch success file")
)

type LineProcess struct {
	filename    string
	processFunc Func
	errCallback ErrCallback

	parallelism  int
	queueTimeout time.Duration
	queueSize    int
	dataQueue    chan string
	once         sync.Once

	doneWG sync.WaitGroup
}

type Func func(line string) error
type ErrCallback func(line string, err error)

//const defaultParallelism = 5
//const defaultQueueSize = 1000
const defaultQueueTimeout = time.Second * 50
const DefaultSuffix = "_success"
const DefaultSeparator = "\t"

func NewLineProcess(filename string, processFunc Func, errCallback ErrCallback) *LineProcess {
	return &LineProcess{
		filename:     filename,
		processFunc:  processFunc,
		errCallback:  errCallback,
		parallelism:  *defaultParallelism,
		queueSize:    *defaultQueueSize,
		queueTimeout: defaultQueueTimeout,
	}
}

func (p *LineProcess) WithParallelism(parallel int) *LineProcess {
	p.parallelism = parallel
	return p
}

func (p *LineProcess) WithQueueTimeout(timeout time.Duration) *LineProcess {
	p.queueTimeout = timeout
	return p
}

func (p *LineProcess) WithQueueSize(size int) *LineProcess {
	p.queueSize = size
	return p
}

func (p *LineProcess) init() error {
	p.dataQueue = make(chan string, p.queueSize)
	return p.startWorker()
}

func (p *LineProcess) Run() error {
	if err := p.init(); err != nil {
		return err
	}
	defer p.close()

	file, err := os.Open(p.filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if err = p.addLine(line); err != nil {
			return err
		}
	}
	return nil
}

func (p *LineProcess) LoadFile(nlines int, separator string, suffix string) error {
	if IsFile(p.filename) {
		if strings.HasSuffix(p.filename, suffix) {
			return nil
		}
		successFlag := p.filename + suffix
		if IsFile(successFlag) && *touchSuccess {
			return nil
		}
		if err := p.RunNLines(nlines, separator); err != nil {
			return err
		}
		if *touchSuccess {
			TouchFile(successFlag)
		}
	} else if IsDir(p.filename) {
		listFiles := GetFileList(p.filename)
		for _, fileName := range listFiles {
			p.filename = fileName
			if err := p.LoadFile(nlines, separator, suffix); err != nil {
				return err
			}
		}
	} else {
		return errors.New("Neither a file nor a folder")
	}
	return nil
}

func (p *LineProcess) RunNLines(nlines int, separator string) error {
	if err := p.init(); err != nil {
		return err
	}
	defer p.close()

	file, err := os.Open(p.filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	// index := 0
	// beginTime := time.Now().Unix()
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		// if index%nlines == 0 {
		// 	curTime := time.Now().Unix()
		// 	fmt.Printf("Process %d lines, Time used %d\n", index, curTime-beginTime)
		// }
		// index++

		if len(lines) == nlines {
			if err = p.addLine(strings.Join(lines, separator)); err != nil {
				glog.Error("Error:", err)
			}
			lines = lines[:0]
		}
	}
	if err = p.addLine(strings.Join(lines, separator)); err != nil {
		return err
	}
	return nil
}

func (p *LineProcess) startWorker() error {
	p.doneWG.Add(p.parallelism)
	for i := 0; i < p.parallelism; i++ {
		go func() {
			defer p.doneWG.Done()
			for line := range p.dataQueue {
				if err := p.processFunc(line); err != nil && p.errCallback != nil {
					p.errCallback(line, err)
				}
			}
		}()
	}
	return nil
}

func (p *LineProcess) addLine(line string) error {
	if line == "" {
		return nil
	}
	tm := time.NewTimer(p.queueTimeout)

	select {
	case p.dataQueue <- line:
		return nil
	case <-tm.C:
		return ErrQueueTimeout
	}
}

func (p *LineProcess) close() {
	// p.once.Do(
	// 	func() {
	// 		close(p.dataQueue)
	// 	})
	close(p.dataQueue)
}

func (p *LineProcess) WaitDone() {
	p.doneWG.Wait()
}
