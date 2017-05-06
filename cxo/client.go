package cxo

import (
	"errors"
	"github.com/evanlinjin/bbs/typ"
	"github.com/skycoin/cxo/node"
	"github.com/skycoin/skycoin/src/cipher"
	"strconv"
	"time"
)

// CXOConfig represents a configuration for CXO Client.
type CXOConfig struct {
	Master bool
	Port   int
}

// NewCXOConfig creates a new CXOConfig.
func NewCXOConfig() *CXOConfig {
	c := CXOConfig{
		Master: false,
		Port:   8998,
	}
	return &c
}

// Client acts as a client for cxo.
type Client struct {
	*CXOConfig
	Client       *node.Client
	UserManager  *typ.UserManager
	BoardManager *typ.BoardManager
}

// NewClient creates a new Client.
func NewClient(conf *CXOConfig) (*Client, error) {

	// Setup cxo client.
	clientConfig := node.NewClientConfig()
	client, e := node.NewClient(clientConfig)
	if e != nil {
		return nil, e
	}
	c := Client{
		CXOConfig:    conf,
		Client:       client,
		UserManager:  typ.NewUserManager(),
		BoardManager: typ.NewBoardManager(conf.Master),
	}
	c.initUsers()
	return &c, nil
}

func (c *Client) initUsers() {
	m := c.UserManager
	if m.Load() != nil {
		m.Clear()
		m.AddNewRandomMaster()
		m.Save()
	}
}

// Launch runs the Client.
func (c *Client) Launch() error {
	if e := c.Client.Start("[::]:" + strconv.Itoa(c.Port)); e != nil {
		return e
	}
	time.Sleep(5 * time.Second)
	c.Client.Execute(func(ct *node.Container) error {
		ct.Register("Board", typ.Board{},
			"BoardThreads", typ.BoardThreads{},
			"Thread", typ.Thread{},
			"Post", typ.Post{})
		return nil
	})
	return nil
}

// Shutdown shutdowns the Client.
func (c *Client) Shutdown() error {
	return c.Client.Close()
}

// SubscribeToBoard subscribes to a board.
func (c *Client) SubscribeToBoard(pk cipher.PubKey) (*typ.BoardConfig, error) {
	bc := &typ.BoardConfig{
		Master:       false,
		PublicKeyStr: pk.Hex(),
	}
	// See if board exists in manager, if not create it.
	if c.BoardManager.AddConfig(bc) != nil {
		var e error
		bc, e = c.BoardManager.GetConfig(pk)
		if e != nil {
			return nil, e
		}
	}
	// Subscribe to board in cxo.
	if c.Client.Subscribe(pk) == false {
		c.BoardManager.RemoveConfig(pk)
		return nil, errors.New("cxo failed to subscribe")
	}
	return bc, nil
}

// UnSubscribeFromBoard unsubscribes from a board.
func (c *Client) UnSubscribeFromBoard(pk cipher.PubKey) bool {
	c.BoardManager.RemoveConfig(pk)
	return c.Client.Unsubscribe(pk)
}

// InjectBoard injects a board with specified BoardConfig.
func (c *Client) InjectBoard(bc *typ.BoardConfig) error {
	// Check if BoardConfig is master.
	if bc.Master == false {
		return errors.New("is not master")
	}
	// Check if BoardConfig already exists in BoardManager.
	if c.BoardManager.HasConfig(bc.PublicKey) == true {
		return errors.New("config already exists")
	}
	// Create Board and BoardThreads for cxo.
	c.Client.Execute(func(ct *node.Container) error {
		root := ct.NewRoot(bc.PublicKey, bc.SecretKey)
		board := typ.NewBoard(bc.Name, bc.URL)
		threads := typ.NewBoardThreads()
		root.Inject(*board, bc.SecretKey)
		root.Inject(*threads, bc.SecretKey)
		return nil
	})
	// Subscribe to board in cxo.
	if c.Client.Subscribe(bc.PublicKey) == false {
		c.BoardManager.RemoveConfig(bc.PublicKey)
		return errors.New("cxo failed to subscribe")
	}
	// Add to BoardManager.
	e := c.BoardManager.AddConfig(bc)
	return e
}

// InjectThread injects a thread to specified board.
func (c *Client) InjectThread(bpk cipher.PubKey, thread *typ.Thread) error {
	// Get BoardConfig.
	bc, e := c.BoardManager.GetConfig(bpk)
	if e != nil {
		return e
	}
	// Check if board is master.
	if bc.Master == false {
		return errors.New("not master")
	}
	// Obtain latest BoardThreads.
	bts, _, e := typ.ObtainLatestBoardThreads(bpk, c.Client)
	if e != nil {
		return e
	}
	// Add thread to BoardThreads.
	if e:= bts.AddThread(bpk, c.Client, thread); e != nil {
		return e
	}
	// Re-Inject BoardThreads to root.
	e = c.Client.Execute(func(ct *node.Container) error {
		r := ct.Root(bpk)
		r.Inject(*bts, bc.SecretKey)
		return nil
	})
	return e
}
