package kea

import (
	"encoding/binary"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"isc.org/stork/api"
	dbops "isc.org/stork/server/database"
	dbmodel "isc.org/stork/server/database/model"

	"google.golang.org/protobuf/proto"
)

// This query uses a nested SELECT and a window function to choose only the
// most-recently-added row (by ctid, which is approximately the row ID but not
// really). Without the added complexity, the query occasionally fails with
//
// ERROR #21000 ON CONFLICT DO UPDATE command cannot affect row a second time.
//
// This error occurs when two rows with duplicate IP addresses and app IDs try
// to insert into the leases table when one row already exists in that table
// with matching IP address and app ID. The second insert fails because the
// first one already updated that row in the lease table.
const updateLeasesTableQuery = `INSERT INTO lease(
  address,
	hw_address,
	client_id,
	valid_lifetime,
	app_id
)
SELECT
    address,
    hw_address,
    client_id,
    valid_lifetime,
    app_id
FROM (
    SELECT
        address,
        hw_address,
        client_id,
        valid_lifetime,
        app_id,
        row_number() OVER (
            PARTITION BY address, app_id ORDER BY ctid DESC
        ) AS rownum
    FROM lease_update
    WHERE app_id = ?
)
WHERE rownum = 1
ON CONFLICT(address, app_id)
	DO UPDATE SET valid_lifetime = EXCLUDED.valid_lifetime;`

type LeaseFile struct {
	file   *os.File
	reader *csv.Reader
	writer *csv.Writer
}

type LeaseFilePbuf struct {
	file       *os.File
	recSizeBuf []byte
	dataBuf    []byte
}

func CreateLeaseFile(name string) (*LeaseFile, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	lf := &LeaseFile{
		file:   f,
		writer: csv.NewWriter(f),
	}
	header := []string{
		"address",
		"hwaddr",
		"client_id",
		"valid_lifetime",
		/*		"expire",
				"subnet_id",
				"fqdn_fwd",
				"fqdn_rev",
				"hostname",
				"state",
				"user_context", */
		"app_id",
	}
	err = lf.writer.Write(header)
	if err != nil {
		f.Close()
		return nil, err
	}
	lf.writer.Flush()
	err = lf.writer.Error()
	if err != nil {
		f.Close()
		return nil, err
	}
	return lf, nil
}

func CreateLeaseFileProtobuf(name string) (*LeaseFilePbuf, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	lf := &LeaseFilePbuf{
		file:       f,
		recSizeBuf: make([]byte, 2),
		dataBuf:    make([]byte, 256),
	}
	return lf, nil
}

func OpenLeaseFile(name string) (*LeaseFile, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	lf := &LeaseFile{
		file:   f,
		reader: csv.NewReader(f),
	}
	_, err = lf.reader.Read()
	if err != nil {
		return nil, err
	}
	return lf, nil
}

func OpenLeaseFilePbuf(name string) (*LeaseFilePbuf, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	lf := &LeaseFilePbuf{
		file:       f,
		recSizeBuf: make([]byte, 2),
		dataBuf:    make([]byte, 256),
	}
	return lf, nil
}

func (lf *LeaseFile) Close() {
	lf.file.Close()
}

func (lf *LeaseFilePbuf) Close() {
	lf.file.Close()
}

func (lf *LeaseFile) CopyToDatabaseCsv(db *dbops.PgDB, appID uint64) error {
	path, err := filepath.Abs(lf.file.Name())
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM lease_update WHERE app_id=?;", appID)
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf("COPY lease_update(address, hw_address, client_id, valid_lifetime, app_id) FROM '%s' WITH CSV HEADER", path))
	if err != nil {
		return err
	}
	_, err = db.Exec(updateLeasesTableQuery, appID)
	if err != nil {
		return err
	}

	return nil
}

func (lf *LeaseFile) CopyToDatabaseNoTx(db *dbops.PgDB, appID uint64) error {
	path, err := filepath.Abs(lf.file.Name())
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM lease_update WHERE app_id=?;", appID)
	if err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	r := csv.NewReader(f)

	// Discard first row (headings)
	_, err = r.Read()
	if err != nil {
		return err
	}
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		_, err = db.Exec("INSERT INTO lease_update (address, hw_address, client_id, valid_lifetime, app_id) VALUES (?, ?, ?, ?, ?);", record[0], record[1], record[2], record[3], record[4])
		if err != nil {
			return err
		}
	}
	_, err = db.Exec(updateLeasesTableQuery, appID)
	if err != nil {
		return err
	}
	return nil
}

func (lf *LeaseFile) CopyToDatabaseExec(db *dbops.PgDB, appID uint64) error {
	path, err := filepath.Abs(lf.file.Name())
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM lease_update WHERE app_id=?;", appID)
	if err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	r := csv.NewReader(f)

	// Discard first row (headings)
	_, err = r.Read()
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	defer tx.Close()
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		_, err = tx.Exec("INSERT INTO lease_update (address, hw_address, client_id, valid_lifetime, app_id) VALUES (?, ?, ?, ?, ?);", record[0], record[1], record[2], record[3], record[4])
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	_, err = tx.Exec(updateLeasesTableQuery, appID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (lf *LeaseFile) CopyToDatabase(db *dbops.PgDB, appID uint64) error {
	path, err := filepath.Abs(lf.file.Name())
	if err != nil {
		fmt.Println("err 1")
		return err
	}
	_, err = db.Exec("DELETE FROM lease_update WHERE app_id=?;", appID)
	if err != nil {
		fmt.Println("err 2")
		return err
	}
	tx, err := db.Begin()
	defer tx.Close()

	cpy, err := tx.Prepare("INSERT INTO lease_update (address, hw_address, client_id, valid_lifetime, app_id) VALUES ($1, $2, $3, $4, $5);")
	if err != nil {
		fmt.Println("err 3")
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		fmt.Println("err 4")
		_ = tx.Rollback()
		return err
	}
	defer f.Close()
	r := csv.NewReader(f)

	// Discard first row (headings)
	_, err = r.Read()
	if err != nil {
		fmt.Println("err 5")
		_ = tx.Rollback()
		return err
	}
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			_ = tx.Rollback()
			fmt.Println("err 6")
			return err
		}
		_, err = cpy.Exec(record[0], record[1], record[2], record[3], record[4])
		if err != nil {
			_ = tx.Rollback()
			fmt.Println("err 7")
			return err
		}
	}
	_, err = tx.Exec(updateLeasesTableQuery, appID)
	if err != nil {
		fmt.Println("err 8")
		_ = tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		fmt.Println("err 9")
		return err
	}
	return nil
}

func (lf *LeaseFilePbuf) CopyToDatabase(db *dbops.PgDB, appID uint64) error {
	path, err := filepath.Abs(lf.file.Name())
	if err != nil {
		fmt.Println("err 1")
		return err
	}
	_, err = db.Exec("DELETE FROM lease_update WHERE app_id=?;", appID)
	if err != nil {
		fmt.Println("err 2")
		return err
	}
	tx, err := db.Begin()
	defer tx.Close()

	cpy, err := tx.Prepare("INSERT INTO lease_update (address, hw_address, client_id, valid_lifetime, app_id) VALUES ($1, $2, $3, $4, $5);")
	if err != nil {
		fmt.Println("err 3")
		return err
	}
	lf.file, err = os.Open(path)
	if err != nil {
		fmt.Println("err 4")
		_ = tx.Rollback()
		return err
	}
	defer lf.file.Close()

	update := api.LeaseUpdate{}
	for {
		err := lf.Read(&update)
		if err == io.EOF {
			break
		}
		if err != nil {
			_ = tx.Rollback()
			//fmt.Println("err 6")
			return err
		}
		_, err = cpy.Exec(update.Address, update.HwAddress, update.ClientId, update.ValidLifetime, update.AppId)
		if err != nil {
			_ = tx.Rollback()
			fmt.Println("err 7")
			return err
		}
	}
	_, err = tx.Exec(updateLeasesTableQuery, appID)
	if err != nil {
		fmt.Println("err 8")
		_ = tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		fmt.Println("err 9")
		return err
	}
	return nil
}

func (lf *LeaseFile) Read() (*dbmodel.LeaseUpdate, error) {
	record, err := lf.reader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil
		}
		return nil, err
	}
	validLifetime, err := strconv.ParseUint(record[3], 10, 32)
	if err != nil {
		return nil, err
	}
	appID, err := strconv.ParseUint(record[4], 10, 64)
	if err != nil {
		return nil, err
	}
	lease := &dbmodel.LeaseUpdate{
		Address:       record[0],
		HWAddress:     nil,
		ClientID:      nil,
		ValidLifetime: uint32(validLifetime),
		AppID:         appID,
	}
	return lease, nil
}

func (lf *LeaseFilePbuf) Read(update *api.LeaseUpdate) error {
	n, err := io.ReadFull(lf.file, lf.recSizeBuf)
	if err != nil {
		return err
	}
	if n != 2 {
		return errors.New("ReadFull didn't read the full buffer?")
	}
	pbufLength := int(binary.BigEndian.Uint16(lf.recSizeBuf))
	if cap(lf.dataBuf) < pbufLength {
		lf.dataBuf = make([]byte, pbufLength)
	}
	// Slice the existing buffer to the right length to avoid allocations.
	pbufData := lf.dataBuf[:pbufLength]
	n, err = io.ReadFull(lf.file, pbufData)
	if n != pbufLength {
		return fmt.Errorf("data file is corrupt: expected to read %d, actually read %d", pbufLength, n)
	}
	pbuferr := proto.Unmarshal(pbufData, update)
	if pbuferr != nil {
		return pbuferr
	}
	return err
}

func (lf *LeaseFile) Write(lease *dbmodel.LeaseUpdate) error {
	record := []string{
		lease.Address,
		"",
		"",
		strconv.FormatUint(uint64(lease.ValidLifetime), 10),
		strconv.FormatUint(lease.AppID, 10),
	}
	err := lf.writer.Write(record)
	if err != nil {
		return err
	}
	return nil
}

func (lf *LeaseFilePbuf) Write(lease *api.LeaseUpdate) error {
	out, err := proto.Marshal(lease)
	if err != nil {
		return err
	}
	if len(out) > (1 << 16) {
		return errors.New("message too long")
	}
	// Use network byte order for maximal standards compliance (even though x86 is little-endian).
	binary.BigEndian.PutUint16(lf.recSizeBuf, uint16(len(out)))
	lf.file.Write(lf.recSizeBuf)
	lf.file.Write(out)
	return nil
}

func (lf *LeaseFile) Flush() error {
	lf.writer.Flush()
	err := lf.writer.Error()
	if err != nil {
		return err
	}
	return nil
}

func (lf *LeaseFilePbuf) Flush() error {
	return lf.file.Sync()
}
