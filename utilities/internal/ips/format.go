package ips

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sort"
)

type FormatOptions struct {
	Host string
	Date string
}

func Format(writer io.Writer, entries []HostEntry, options FormatOptions) error {
	if _, err := fmt.Fprintln(writer, "# gofips fixed IP assignments"); err != nil {
		return err
	}
	if options.Host != "" {
		if _, err := fmt.Fprintf(writer, "# exported from UDM at %s\n", options.Host); err != nil {
			return err
		}
	}
	if options.Date != "" {
		if _, err := fmt.Fprintf(writer, "# date: %s\n", options.Date); err != nil {
			return err
		}
	}

	if len(entries) == 0 {
		if _, err := fmt.Fprintln(writer, "#"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(writer, "# No fixed IP assignments found."); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(writer, "# Example:"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(writer, "# host mydevice {"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(writer, "#     hardware ethernet aa:bb:cc:dd:ee:ff;"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(writer, "#     fixed-address 192.168.1.10;"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(writer, "# }"); err != nil {
			return err
		}
		return nil
	}

	sorted := make([]HostEntry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		return ipToUint32(sorted[i].IP) < ipToUint32(sorted[j].IP)
	})

	for _, entry := range sorted {
		if _, err := fmt.Fprintln(writer); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(writer, "host %s {\n", entry.Hostname); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(writer, "    hardware ethernet %s;\n", entry.MAC); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(writer, "    fixed-address %s;\n", entry.IP); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(writer, "}"); err != nil {
			return err
		}
	}

	return nil
}

func ipToUint32(ipString string) uint32 {
	parsed := net.ParseIP(ipString)
	if parsed == nil {
		return 0
	}
	ipv4 := parsed.To4()
	if ipv4 == nil {
		return 0
	}
	return binary.BigEndian.Uint32(ipv4)
}
