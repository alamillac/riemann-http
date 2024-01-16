package asn

import (
  "fmt"
  "log"
  "net"
  "strings"
)

func fetchASNFromDNS(ip string) (string, error) {
  // Invert the IP address
  reversedIP, err := reverseIP(ip)
  if err != nil {
    return "", err
  }

  // Perform a DNS query to get the ASN
  log.Printf("Lookup to: %s", reversedIP)
  answers, err := net.LookupTXT(reversedIP)
  if err != nil {
    return "", err
  }

  if len(answers) == 0 {
    return "Unknown", nil
  }

  asn, _, _, _, err := parseASNData(answers[0])
  if err != nil {
    return "", err
  }

  return asn, nil
}

func parseASNData(data string) (asn, network, country, createDate string, err error) {
  // ASN, network, country, name, createDate
  // 8151 | 2806:108e:13::/48 | MX | lacnic | 2011-03-01
  // 397630 | 154.83.10.0/24 | SC | afrinic | 2013-07-24

  log.Printf("Received: %s", data)
  matches := strings.Split(data, "|")
  if len(matches) < 5 {
    return "", "", "", "", fmt.Errorf("formato no válido")
  }

  // Asignar los valores extraídos
  asn = strings.TrimSpace(matches[0])
  network = strings.TrimSpace(matches[1])
  country = strings.TrimSpace(matches[2])
  createDate = strings.TrimSpace(matches[4])

  log.Printf("ASN: %s, Red: %s, País: %s, Fecha de Creación: %s", asn, network, country, createDate)
  return asn, network, country, createDate, nil
}

func reverseIP(ip string) (string, error) {
  parsedIP := net.ParseIP(ip)
  if parsedIP == nil {
    return "", fmt.Errorf("invalid IP address")
  }

  // Soporte para IPv4
  if ip4 := parsedIP.To4(); ip4 != nil {
    return fmt.Sprintf("%d.%d.%d.%d.origin.asn.cymru.com", ip4[3], ip4[2], ip4[1], ip4[0]), nil
  }

  // Soporte para IPv6
  if ip6 := parsedIP.To16(); ip6 != nil {
    var reversedIP6 string
    for i := len(ip6) - 1; i >= 0; i-- {
      reversedIP6 += reverseHex(ip6[i])
      if i > 0 {
        reversedIP6 += "."
      }
    }
    return reversedIP6 + ".origin6.asn.cymru.com", nil
  }

  return "", fmt.Errorf("unknown IP address format")
}

func reverseHex(hex byte) string { 
  return fmt.Sprintf("%x", hex & 0x0F) + "." + fmt.Sprintf("%x", (hex & 0xF0) >> 4)
}
