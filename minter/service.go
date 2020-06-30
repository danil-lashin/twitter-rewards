package minter

import (
	"github.com/MinterTeam/minter-go-sdk/api"
	"github.com/MinterTeam/minter-go-sdk/transaction"
	"github.com/MinterTeam/minter-go-sdk/wallet"
	"github.com/danil-lashin/twitter-rewards/config"
	"github.com/danil-lashin/twitter-rewards/helpers"
	"math/big"
	"time"
)

type Service struct {
	masterAddress string
	privateKey    string
	coin          string
	value         *big.Int

	client   *api.Api
	sendChan chan SendJob
}

type SendJob struct {
	To       string
	Callback func()
}

func NewService(cfg *config.Config) *Service {
	value, _ := big.NewInt(0).SetString(cfg.Amount, 10)

	seed, _ := wallet.Seed(cfg.MinterMnemonic)
	privateKey, _ := wallet.PrivateKeyBySeed(seed)
	publicKey, _ := wallet.PublicKeyByPrivateKey(privateKey)
	address, _ := wallet.AddressByPublicKey(publicKey)

	service := &Service{
		masterAddress: address,
		privateKey:    privateKey,
		coin:          cfg.Coin,
		value:         value,
		client:        api.NewApi(cfg.MinterNodeUrl),
		sendChan:      make(chan SendJob, 100),
	}

	service.run()

	return service
}

func (s *Service) run() {
	go func() {
		ticker := time.NewTicker(time.Second * 5)

		for range ticker.C {
			var jobs []SendJob
			done := false
			for !done {
				select {
				case job := <-s.sendChan:
					jobs = append(jobs, job)
				default:
					done = true
				}
			}

			if len(jobs) < 1 {
				continue
			}

			nonce, err := s.client.Nonce(s.masterAddress)
			if err != nil {
				panic(err)
			}

			data := transaction.NewMultisendData()
			for _, job := range jobs {
				data.AddItem(*transaction.NewMultisendDataItem().SetCoin(s.coin).SetValue(helpers.BipToPip(s.value)).MustSetTo(job.To))
			}

			tx, err := transaction.NewBuilder(transaction.MainNetChainID).NewTransaction(data)
			if err != nil {
				panic(err)
			}

			tx.SetNonce(nonce).SetGasPrice(1).SetGasCoin("BIP").SetSignatureType(transaction.SignatureTypeSingle)
			signedTx, _ := tx.Sign(s.privateKey)

			result, err := s.client.SendTransaction(signedTx)
			if err != nil {
				panic(err)
			}

			for {
				time.Sleep(time.Second)
				_, err := s.client.Transaction(result.Hash)
				if err != nil {
					break
				}
			}

			for _, job := range jobs {
				go job.Callback()
			}
		}
	}()
}

func (s *Service) AddJob(address string, callback func()) {
	s.sendChan <- SendJob{
		To:       address,
		Callback: callback,
	}
}
