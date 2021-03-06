package main

import (
	"testing"
	"time"
	// "fmt"
	. "github.com/franela/goblin"
)

func createDummyOrders(n int64) [6]Order {
	return [6]Order{
				BuyLimit{
					bid: 10.05, 
					BaseOrder:BaseOrder{
						actor: "Bob", timecreated: time.Now().Unix() + n, 
						intent: "BUY", shares: 100, state: "OPEN",
					},
				},
				BuyMarket{
					BaseOrder:BaseOrder{
						actor: "Tim", timecreated: time.Now().Unix() + n,
						intent: "BUY", shares: 100, state: "OPEN",
					},
				},
				BuyLimit{
					bid: 10.00, 
					BaseOrder:BaseOrder{
						actor: "Gary", timecreated: time.Now().Unix() + n,
						intent: "BUY", shares: 100, state: "OPEN",
					},
				},
				SellMarket{
					BaseOrder:BaseOrder{
						actor: "Terry", timecreated: time.Now().Unix() + n,
						intent: "SELL", shares: 100, state: "OPEN",
					},
				},
				SellLimit{
					ask: 10.10, 
					BaseOrder:BaseOrder{
						actor: "Larry", timecreated: time.Now().Unix() + n,
						intent: "SELL", shares: 100, state: "OPEN",
					},
				},
				SellMarket{
					BaseOrder:BaseOrder{
						actor: "Sam", timecreated: time.Now().Unix() + n,
						intent: "SELL", shares: 100, state: "OPEN",
					},
				},
			}
}

func Test(t *testing.T){
	g := Goblin(t)

	g.Describe("Orders", func(){

		var orders [6]Order
		var moreOrders [6]Order

		g.BeforeEach(func(){
			orders = createDummyOrders(0)
			moreOrders = createDummyOrders(1)
		})

		g.Describe("Order Interface", func(){

			g.Describe("price method", func(){
				g.It("should equal the bid on a BuyLimit", func(){
					g.Assert(orders[0].price()).Equal(10.05)
				})

				g.It("should equal the ask on a SellLimit", func(){
					g.Assert(orders[4].price()).Equal(10.10)
				})

				g.It("should equal nearly infinity on BuyMarket", func(){
					g.Assert(orders[1].price()).Equal(1000000.00)
				})

				g.It("should equal 0 on SellMarket", func(){
					g.Assert(orders[5].price()).Equal(0.00)
				})
			})

			g.Describe("lookup method", func(){
				g.It("should equal actor + createtime", func(){
					g.Assert(orders[0].lookup()[:3]).Equal("Bob")
				})
			})

			g.Describe("getOrder method", func(){
				g.It("should provide access to the embeded order struct", func(){
					g.Assert(orders[0].getOrder().shares).Equal(100)
				})
			})
		})

		g.Describe("OrderBook", func(){

			g.It("should add orders to the correct queues and hashes", func(){
				orderBook := NewOrderBook()

				for i := 0; i < len(orders); i++ {
					orderBook.add(orders[i])
				}

				g.Assert(orderBook.buyQueue.Dequeue().Value).Equal(1000000.00)
				g.Assert(orderBook.buyQueue.Dequeue().Value).Equal(10.05)
				g.Assert(orderBook.buyQueue.Dequeue().Value).Equal(10.00)
				
				g.Assert(orderBook.sellQueue.Dequeue().Value).Equal(0.00)
				g.Assert(orderBook.sellQueue.Dequeue().Value).Equal(0.00)
				g.Assert(orderBook.sellQueue.Dequeue().Value).Equal(10.10)
			})

			g.It("should fill the highest priority orders until no more can be filled", func(){

				orderBook := NewOrderBook()

				for i :=0; i < len(orders); i++ {
					orderBook.add(orders[i])
				}

				// filling orders will dequeue filled orders,
				// so expect further down the line orders when dequeueing
				orderBook.run()
				g.Assert(orderBook.buyQueue.Dequeue().Value).Equal(10.00)
				g.Assert(orderBook.sellQueue.Dequeue().Value).Equal(10.10)
			})

			g.It("should work with repeated calls to add and run", func(){
				orderBook := NewOrderBook()

				for i :=0; i < len(orders); i++ {
					orderBook.add(orders[i])
				}

				orderBook.run()

				orderBook.add(orders[1])

				orderBook.run()

				g.Assert(orderBook.buyQueue.Dequeue().Value).Equal(10.00)
				g.Assert(orderBook.sellQueue.Dequeue() == nil).Equal(true)
			})

			g.It("should call tradeHandler with matched orders", func(){
				
				orderBook := NewOrderBook()

				orderBook.setTradeHandler(func (o Order) {
					if o.getOrder().intent == "BUY"{
						g.Assert(o.price()).Equal(10.05)
					} else {
						g.Assert(o.price()).Equal(0.00)
					}
				})

				orderBook.add(orders[0])
				orderBook.add(orders[3])
				orderBook.run()
			})

			g.It("should partially fill orders when the share numbers dont match", func(){
				
			})
			
		})

	})
}

// BENCHMARKS
var benchOrders = createDummyOrders(0)

var result float64

func BenchmarkOrderBookRun(b *testing.B){

	orderBook := NewOrderBook()

	for i :=0; i < len(benchOrders); i++ {
		orderBook.add(benchOrders[i])
	}

	// filling orders will dequeue filled orders,
	// so expect further down the line orders when dequeueing
	orderBook.run()
	result = orderBook.buyQueue.Dequeue().Value
}