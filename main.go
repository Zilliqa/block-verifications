package main

import (
	"container/list"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/Zilliqa/gozilliqa-sdk/core"
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	verifier2 "github.com/Zilliqa/gozilliqa-sdk/verifier"
)

func main() {
	p := provider.NewProvider("https://api.zilliqa.com")
	initDsComm, _ := p.GetCurrentDSComm()
	log.Println("current tx block num: " + initDsComm.CurrentTxEpoch)
	log.Println("current ds block num: " + initDsComm.CurrentDSEpoch)
	log.Println("current ds comm: ", initDsComm.DSComm)
	log.Println("number of ds guard: ", initDsComm.NumOfDSGuard)
	currentTxBlockNum, _ := strconv.ParseUint(initDsComm.CurrentTxEpoch, 10, 64)
	currentDsBlockNum, _ := strconv.ParseUint(initDsComm.CurrentDSEpoch, 10, 64)
	verifier := &verifier2.Verifier{NumOfDsGuard: initDsComm.NumOfDSGuard}

	for {
		latestTxBlock, _ := p.GetLatestTxBlock()
		log.Println("wait current tx block got generated")
		latestTxBlockNum, _ := strconv.ParseUint(latestTxBlock.Header.BlockNum, 10, 64)
		log.Printf("latest tx block num is: %d, current tx block num is: %d\n", latestTxBlockNum, currentTxBlockNum)
		if latestTxBlockNum > currentTxBlockNum {
			break
		}
		time.Sleep(time.Second * 5)
	}

	dsComm := list.New()
	for _, ds := range initDsComm.DSComm {
		dsComm.PushBack(core.PairOfNode{
			PubKey: ds,
		})
	}

	dst, _ := p.GetDsBlockVerbose(initDsComm.CurrentDSEpoch)
	dsBlock := core.NewDsBlockFromDsBlockT(dst)
	initDsBlock, _ := json.Marshal(dsBlock)
	log.Println("init ds block raw: ")
	log.Println(string(initDsBlock))

	tst, _ := p.GetTxBlockVerbose(initDsComm.CurrentTxEpoch)
	txBlock := core.NewTxBlockFromTxBlockT(tst)
	initTxBlock, _ := json.Marshal(txBlock)
	log.Println("init tx block raw: ")
	log.Println(string(initTxBlock))

	err := verifier.VerifyTxBlock(txBlock, dsComm)
	if err != nil {
		log.Fatalln("verify init tx block error: " + err.Error())
	}
	log.Println("verify init tx block succeed")

	for {
		log.Println("get latest block")
		latestTxBlock, _ := p.GetLatestTxBlock()
		latest, _ := strconv.ParseUint(latestTxBlock.Header.BlockNum, 10, 64)
		if latest > currentTxBlockNum {
			currentTxBlockNum++
			// before handle tx block, check ds block first
			var txBlockT *core.TxBlockT
		T:
			for {
				txBlockT, _ = p.GetTxBlockVerbose(strconv.FormatUint(currentTxBlockNum, 10))
				if txBlockT.Header.DSBlockNum == "18446744073709551615" {
					time.Sleep(time.Second)
					log.Println("re-get block data")
					goto T
				}
				goto G

			}
		G:
			dsBlockNum, _ := strconv.ParseUint(txBlockT.Header.DSBlockNum, 10, 64)

			if dsBlockNum > currentDsBlockNum {
				currentDsBlockNum++
				dsBlockT, _ := p.GetDsBlockVerbose(strconv.FormatUint(dsBlockNum, 10))
				dsBlock := core.NewDsBlockFromDsBlockT(dsBlockT)
				log.Println("ds block, block number = ", dsBlock.BlockHeader.BlockNum)
				newDsComm, err := verifier.VerifyDsBlock(dsBlock, dsComm)
				if err == nil {
					log.Printf("verify ds block %d succeed\n", dsBlockNum)
				} else {
					log.Fatalf("verify ds block %d failed\n", dsBlockNum)
				}
				dsComm = newDsComm
			}

			log.Println("tx block, block number = ", txBlockT.Header.BlockNum)
			txBlock, _ := json.Marshal(core.NewTxBlockFromTxBlockT(txBlockT))
			log.Println(string(txBlock))
			err := verifier.VerifyTxBlock(core.NewTxBlockFromTxBlockT(txBlockT), dsComm)
			if err == nil {
				log.Printf("verify tx block %d succeed\n", currentTxBlockNum)
			} else {
				log.Fatalf("verify tx block %d failed\n", currentTxBlockNum)
			}
		} else {
			log.Println("sleep to wait new block")
			time.Sleep(time.Second)
		}
	}

}
