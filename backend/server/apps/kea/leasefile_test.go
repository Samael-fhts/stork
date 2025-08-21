package kea

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"testing"

	log "github.com/sirupsen/logrus"
	"isc.org/stork/api"
	dbmodel "isc.org/stork/server/database/model"
	dbtest "isc.org/stork/server/database/test"
)

func BenchmarkLeaseFileLoad(b *testing.B) {
	log.SetLevel(log.FatalLevel)
	log.SetOutput(io.Discard)
	benchmarks := []struct {
		name  string
		count int
	}{
		{
			name:  "10,500 leases",
			count: 1500,
		},
		{
			name:  "35,000 leases",
			count: 5000,
		},
		{
			name:  "140,000 leases",
			count: 20000,
		},
		{
			name:  "700,000 leases",
			count: 100000,
		},
		{
			name:  "7M leases",
			count: 1000000,
		},
	}

	for bi := range benchmarks {
		count := benchmarks[bi].count
		b.Run(benchmarks[bi].name, func(b *testing.B) {
			var wg sync.WaitGroup
			for i := range 7 {
				wg.Add(1)
				go func(wg *sync.WaitGroup, count int, appID uint64) {
					defer wg.Done()
					lf, err := CreateLeaseFile(fmt.Sprintf("./leasefile%d.csv", appID))
					if err != nil {
						b.Logf("failed to create new lease file: %s", err)
						return
					}

					for i := range count {
						ipv4 := net.IPv4(uint8(i>>24), uint8(i>>16), uint8(i>>8), uint8(i))
						lease := &dbmodel.LeaseUpdate{
							Address:       ipv4.String(),
							ValidLifetime: 3600,
							AppID:         appID,
						}
						err = lf.Write(lease)
						if err != nil {
							b.Logf("failed to write a lease to a lease file: %s", err)
							return
						}
					}
					err = lf.Flush()
					if err != nil {
						b.Logf("failed to flush the lease file: %s", err)
						return
					}
					lf.Close()
				}(&wg, count, uint64(i))

				defer os.Remove(fmt.Sprintf("./leasefile%d.csv", i))
			}

			wg.Wait()

			db, _, teardown := dbtest.SetupDatabaseTestCase(b)
			defer teardown()

			b.ResetTimer()
			for b.Loop() {
				for i := range 7 {
					wg.Add(1)
					go func(wg *sync.WaitGroup, appID uint64) {
						defer wg.Done()
						lf, err := OpenLeaseFile(fmt.Sprintf("./leasefile%d.csv", appID))
						if err != nil {
							b.Logf("failed to open a lease file: %s", err)
							return
						}
						err = lf.CopyToDatabase(db, appID)
						if err != nil {
							b.Logf("failed to copy lease file to the database: %s", err)
							return
						}
						lf.Close()
						// fmt.Printf("Finished copying data for appID=%d\n", appID)
					}(&wg, uint64(i))
				}
				wg.Wait()

				// count, _ := db.Model(&dbmodel.Lease{}).Count()
				// fmt.Printf("COUNT IS %d\n", count)
			}
		})
	}
}

func BenchmarkLeaseFileLoadProtobuf(b *testing.B) {
	log.SetLevel(log.FatalLevel)
	log.SetOutput(io.Discard)
	benchmarks := []struct {
		name  string
		count int
	}{
		{
			name:  "10,500 leases",
			count: 1500,
		},
		{
			name:  "35,000 leases",
			count: 5000,
		},
		{
			name:  "140,000 leases",
			count: 20000,
		},
		{
			name:  "700,000 leases",
			count: 100000,
		},
		{
			name:  "7M leases",
			count: 1000000,
		},
	}

	for bi := range benchmarks {
		count := benchmarks[bi].count
		b.Run(benchmarks[bi].name, func(b *testing.B) {
			var wg sync.WaitGroup
			for i := range 7 {
				wg.Add(1)
				go func(wg *sync.WaitGroup, count int, appID uint64) {
					defer wg.Done()
					lf, err := CreateLeaseFileProtobuf(fmt.Sprintf("./leasefile%d.bin", appID))
					if err != nil {
						b.Logf("failed to create new lease file: %s", err)
						return
					}

					for i := range count {
						ipv4 := net.IPv4(uint8(i>>24), uint8(i>>16), uint8(i>>8), uint8(i))
						lease := &api.LeaseUpdate{
							Address:       ipv4.String(),
							ValidLifetime: 3600,
							AppId:         appID,
						}
						err = lf.Write(lease)
						if err != nil {
							b.Logf("failed to write a lease to a lease file: %s", err)
							return
						}
					}
					err = lf.Flush()
					if err != nil {
						b.Logf("failed to flush the lease file: %s", err)
						return
					}
					lf.Close()
				}(&wg, count, uint64(i))

				defer os.Remove(fmt.Sprintf("./leasefile%d.bin", i))
			}

			wg.Wait()

			db, _, teardown := dbtest.SetupDatabaseTestCase(b)
			defer teardown()

			b.ResetTimer()
			for b.Loop() {
				for i := range 7 {
					wg.Add(1)
					go func(wg *sync.WaitGroup, appID uint64) {
						defer wg.Done()
						lf, err := OpenLeaseFilePbuf(fmt.Sprintf("./leasefile%d.bin", appID))
						if err != nil {
							b.Logf("failed to open a lease file: %s", err)
							return
						}
						err = lf.CopyToDatabase(db, appID)
						if err != nil {
							b.Logf("failed to copy lease file to the database: %s", err)
							return
						}
						lf.Close()
						// fmt.Printf("Finished copying data for appID=%d\n", appID)
					}(&wg, uint64(i))
				}
				wg.Wait()

				// count, _ := db.Model(&dbmodel.Lease{}).Count()
				// fmt.Printf("COUNT IS %d\n", count)
			}
		})
	}
}
