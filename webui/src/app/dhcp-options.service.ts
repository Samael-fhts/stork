import { Injectable } from '@angular/core'
import { DhcpOptionDef } from './dhcp-option-def'
import stdDhcpv4OptionDefsRaw from './std-dhcpv4-option-defs.json'
import stdDhcpv6OptionDefsRaw from './std-dhcpv6-option-defs.json'
import { Observable } from 'rxjs'

/**
 * Converts the raw JSON data into the structures used by the application.
 */
const stdDhcpv4OptionDefs: DhcpOptionDef[] = stdDhcpv4OptionDefsRaw.map((raw) => ({
    array: raw.array,
    code: raw.code,
    name: raw.name,
    encapsulate: raw.encapsulate,
    optionType: raw.type,
    space: raw.space,
    recordTypes: raw['record-types']?.split(',').map((t) => t.trim()) ?? [],
}))

const stdDhcpv6OptionDefs: DhcpOptionDef[] = stdDhcpv6OptionDefsRaw.map((raw) => ({
    array: raw.array,
    code: raw.code,
    name: raw.name,
    encapsulate: raw.encapsulate,
    optionType: raw.type,
    space: raw.space,
    recordTypes: raw['record-types']?.split(',').map((t) => t.trim()) ?? [],
}))

/**
 * An interface to a DHCP option description.
 *
 * It is used to define a list of standard DHCP options.
 */
export interface DhcpOptionListItem {
    label: string
    value: number
}

/**
 * A service exposing a list of DHCP options with their mapping between
 * option codes and friendly names.
 *
 * A full list of options returned by this service can be used in the
 * forms containing a list of available options. In the components
 * that display configured options it is useful to find option name by
 * the option code value. This service provides such a capability.
 */
@Injectable({
    providedIn: 'root',
})
export class DhcpOptionsService {
    /**
     * Defines a list of the user-configurable standard DHCPv4 options.
     */
    private configurableDHCPv4OptionCodes: number[] = [
        // 1, // '(1) Subnet Mask'
        2, // '(2) Time Offset'
        3, // '(3) Router'
        4, // '(4) Time Server'
        5, // '(5) Name Server'
        6, // '(6) Domain Server'
        7, // '(7) Log Server'
        8, // '(8) Quotes Server'
        9, // '(9) LPR Server'
        10, // '(10) Impress Server'
        11, // '(11) RLP Server'
        // 12, // '(12) Hostname'
        13, // '(13) Boot File Size'
        14, // '(14) Merit Dump File'
        15, // '(15) Domain Name'
        16, // '(16) Swap Server'
        17, // '(17) Root Path'
        18, // '(18) Extension File'
        19, // '(19) Forward On/Off'
        20, // '(20) SrcRte On/Off'
        21, // '(21) Policy Filter'
        22, // '(22) Max DG Assembly'
        23, // '(23) Default IP TTL'
        24, // '(24) MTU Timeout'
        25, // '(25) MTU Plateau'
        26, // '(26) MTU Interface'
        27, // '(27) MTU Subnet'
        28, // '(28) Broadcast Address'
        29, // '(29) Mask Discovery'
        30, // '(30) Mask Supplier'
        31, // '(31) Router Discovery'
        32, // '(32) Router Request'
        33, // '(33) Static Route'
        34, // '(34) Trailers'
        35, // '(35) ARP Timeout'
        36, // '(36) Ethernet'
        37, // '(37) Default TCP TTL'
        38, // '(38) Keepalive Time'
        39, // '(39) Keepalive Data'
        40, // '(40) NIS Domain'
        41, // '(41) NIS Servers'
        42, // '(42) NTP Servers'
        43, // '(43) Vendor Specific'
        44, // '(44) NETBIOS Name Srv'
        45, // '(45) NETBIOS Dist Srv'
        46, // '(46) NETBIOS Node Type'
        47, // '(47) NETBIOS Scope'
        48, // '(48) X Window Font'
        49, // '(49) X Window Manager'
        // 50, // '(50) Address Request'
        // 51, // '(51) Address Time'
        52, // '(52) Overload'
        // 53, // '(53) DHCP Msg Type'
        54, // '(54) DHCP Server Id'
        // 55, // '(55) Parameter List'
        56, // '(56) DHCP Message'
        57, // '(57) DHCP Max Msg Size'
        // 58, // '(58) Renewal Time'
        // 59, // '(59) Rebinding Time'
        60, // '(60) Class Id'
        // 61, // '(61) Client Id'
        62, // '(62) NetWare/IP Domain'
        63, // '(63) NetWare/IP Option'
        64, // '(64) NIS-Domain-Name'
        65, // '(65) NIS-Server-Addr'
        66, // '(66) Server-Name'
        67, // '(67) Bootfile-Name'
        68, // '(68) Home-Agent-Addrs'
        69, // '(69) SMTP-Server'
        70, // '(70) POP3-Server'
        71, // '(71) NNTP-Server'
        72, // '(72) WWW-Server'
        73, // '(73) Finger-Server'
        74, // '(74) IRC-Server'
        75, // '(75) StreetTalk-Server'
        76, // '(76) STDA-Server'
        77, // '(77) User-Class'
        78, // '(78) Directory Agent'
        79, // '(79) Service Scope'
        // 80, // '(80) Rapid Commit'
        // 81, // '(81) Client FQDN'
        // 82, // '(82) Relay Agent Information'
        // 83, // '(83) iSNS'
        // 84, // '(84) REMOVED/Unassigned'
        85, // '(85) NDS Servers'
        86, // '(86) NDS Tree Name'
        87, // '(87) NDS Context'
        88, // '(88) BCMCS Controller Domain Name list'
        89, // '(89) BCMCS Controller IPv(4) address option'
        // 90, // '(90) Authentication'
        // 91, // '(91) client-last-transaction-time option'
        // 92, // '(92) associated-ip option'
        93, // '(93) Client System'
        94, // '(94) Client NDI'
        // 95, // '(95) LDAP'
        // 96, // '(96) REMOVED/Unassigned'
        97, // '(97) UUID/GUID'
        98, // '(98) User-Auth'
        99, // '(99) GEOCONF_CIVIC'
        100, // '(100) PCode'
        101, // '(101) TCode'
        // 102 - 107, // '(102-107) REMOVED/Unassigned'
        108, // '(108) IPv6-Only Preferred'
        // 109, // '(109) OPTION_DHCP4O6_S46_SADDR'
        // 110, // '(110) REMOVED/Unassigned'
        // 111, // '(111) Unassigned'
        112, // '(112) Netinfo Address'
        113, // '(113) Netinfo Tag'
        114, // '(114) DHCP Captive-Portal'
        // 115, // '(115) REMOVED/Unassigned'
        116, // '(116) Auto-Config'
        117, // '(117) Name Service Search'
        // 118, // '(118) Subnet Selection Option'
        119, // '(119) Domain Search'
        // 120, // '(120) SIP Servers DHCP Option'
        // 121, // '(121) Classless Static Route Option'
        // 122, // '(122) CCC'
        // 123, // '(123) GeoConf Option'
        124, // '(124) V-I Vendor Class'
        125, // '(125) V-I Vendor-Specific Information'
        // 126, // '(126) Removed/Unassigned'
        // 127, // '(127) Removed/Unassigned'
        // 128, // '(128) PXE - undefined (vendor specific)'
        // 128, // '(128) Etherboot signature. 6 bytes: E4:45:74:68:00:00'
        // 128, // '(128) DOCSIS full security server IP address'
        // 128, // '(128) TFTP Server IP address (for IP Phone software load)'
        // 129, // '(129) PXE - undefined (vendor specific)'
        // 129, // '(129) Kernel options. Variable length string'
        // 129, // '(129) Call Server IP address'
        // 130, // '(130) PXE - undefined (vendor specific)'
        // 130, // '(130) Ethernet interface. Variable length string.'
        // 130, // '(130) Discrimination string (to identify vendor)'
        // 131, // '(131) PXE - undefined (vendor specific)'
        // 131, // '(131) Remote statistics server IP address'
        // 132, // '(132) PXE - undefined (vendor specific)'
        // 132, // '(132) IEEE 802.1Q VLAN ID'
        // 133, // '(133) PXE - undefined (vendor specific)'
        // 133, // '(133) IEEE 802.1D/p Layer 2 Priority'
        // 134, // '(134) PXE - undefined (vendor specific)'
        // 134, // '(134) Diffserv Code Point (DSCP) for VoIP signalling and media streams'
        // 135, // '(135) PXE - undefined (vendor specific)'
        // 135, // '(135) HTTP Proxy for phone-specific applications'
        136, // '(136) OPTION_PANA_AGENT'
        137, // '(137) OPTION_V4_LOST'
        138, // '(138) OPTION_CAPWAP_AC_V4'
        // 139, // '(139) OPTION-IPv4_Address-MoS'
        // 140, // '(140) OPTION-IPv4_FQDN-MoS'
        141, // '(141) SIP UA Configuration Service Domains'
        // 142, // '(142) OPTION-IPv4_Address-ANDSF'
        // 143, // '(143) OPTION_V4_SZTP_REDIRECT'
        // 144, // '(144) GeoLoc'
        // 145, // '(145) FORCERENEW_NONCE_CAPABLE'
        146, // '(146) RDNSS Selection'
        // 147, // '(147) OPTION_V4_DOTS_RI'
        // 148, // '(148) OPTION_V4_DOTS_ADDRESS'
        // 149, // '(149) Unassigned'
        // 150, // '(150) TFTP server address'
        // 150, // '(150) Etherboot'
        // 150, // '(150) GRUB configuration path name'
        // 151, // '(151) status-code'
        // 152, // '(152) base-time'
        // 153, // '(153) start-time-of-state'
        // 154, // '(154) query-start-time'
        // 155, // '(155) query-end-time'
        // 156, // '(156) dhcp-state'
        // 157, // '(157) data-source'
        // 158, // '(158) OPTION_V(4)_PCP_SERVER'
        // 159, // '(159) OPTION_V(4)_PORTPARAMS'
        // 160, // '(160) Unassigned'
        // 161, // '(161) OPTION_MUD_URL_V4'
        // 162 - 174, // '(162-174) Unassigned'
        // 175, // '(175) Etherboot (Tentatively Assigned - 2005-06-23)'
        // 176, // '(176) IP Telephone (Tentatively Assigned - 2005-06-23)'
        // 177, // '(177) Etherboot (Tentatively Assigned - 2005-06-23)'
        // 177, // '(177) PacketCable and CableHome (replaced by 122)'
        // 178 - 207, // '(178-207) Unassigned'
        // 208, // '(208) PXELINUX Magic'
        // 209, // '(209) Configuration File'
        // 210, // '(210) Path Prefix'
        // 211, // '(211) Reboot Time'
        212, // '(212) OPTION_6RD'
        213, // '(213) OPTION_V4_ACCESS_DOMAIN'
        // 214 - 219, // '(214-219) Unassigned'
        // 220, // '(220) Subnet Allocation Option'
        // 221, // '(221) Virtual Subnet Selection (VSS) Option'
        // 222 - 223, // '(222-223) Unassigned'
        // 224 - 254, // '(224-254) Reserved (Private Use)'
    ]

    /**
     * Defines a list of the configurable standard DHCPv4 options.
     *
     * Commented out options are not configurable by a user.
     */
    private _dhcpv4Options: DhcpOptionListItem[]

    /**
     * Indexes the standard DHCPv4 options by option code for faster lookup.
     */
    private _dhcpv4OptionsByCode: Map<number, DhcpOptionListItem>

    /**
     * Defines a list of the user-configurable standard DHCPv6 options.
     */
    private configurableDHCPv6OptionCodes: number[] = [
        // 0, // '(0) Reserved'
        // 1, // '(1) OPTION_CLIENTID'
        // 2, // '(2) OPTION_SERVERID'
        // 3, // '(3) OPTION_IA_NA'
        // 4, // '(4) OPTION_IA_TA'
        // 5, // '(5) OPTION_IAADDR'
        // 6, // '(6) OPTION_ORO'
        7, // '(7) OPTION_PREFERENCE'
        // 8, // '(8) OPTION_ELAPSED_TIME'
        // 9, // '(9) OPTION_RELAY_MSG'
        // 10, // '(10) Unassigned'
        // 11, // '(11) OPTION_AUTH'
        12, // '(12) OPTION_UNICAST'
        // 13, // '(13) OPTION_STATUS_CODE'
        // 14, // '(14) OPTION_RAPID_COMMIT'
        // 15, // '(15) OPTION_USER_CLASS'
        // 16, // '(16) OPTION_VENDOR_CLASS'
        // 17, // '(17) OPTION_VENDOR_OPTS'
        // 18, // '(18) OPTION_INTERFACE_ID'
        // 19, // '(19) OPTION_RECONF_MSG'
        // 20, // '(20) OPTION_RECONF_ACCEPT'
        21, // '(21) OPTION_SIP_SERVER_D'
        22, // '(22) OPTION_SIP_SERVER_A'
        23, // '(23) OPTION_DNS_SERVERS'
        24, // '(24) OPTION_DOMAIN_LIST'
        // 25, // '(25) OPTION_IA_PD'
        // 26, // '(26) OPTION_IAPREFIX'
        27, // '(27) OPTION_NIS_SERVERS'
        28, // '(28) OPTION_NISP_SERVERS'
        29, // '(29) OPTION_NIS_DOMAIN_NAME'
        30, // '(30) OPTION_NISP_DOMAIN_NAME'
        31, // '(31) OPTION_SNTP_SERVERS'
        32, // '(32) OPTION_INFORMATION_REFRESH_TIME'
        33, // '(33) OPTION_BCMCS_SERVER_D'
        34, // '(34) OPTION_BCMCS_SERVER_A'
        // 35, // '(35) Unassigned'
        36, // '(36) OPTION_GEOCONF_CIVIC'
        37, // '(37) OPTION_REMOTE_ID'
        38, // '(38) OPTION_SUBSCRIBER_ID'
        39, // '(39) OPTION_CLIENT_FQDN'
        40, // '(40) OPTION_PANA_AGENT'
        41, // '(41) OPTION_NEW_POSIX_TIMEZONE'
        42, // '(42) OPTION_NEW_TZDB_TIMEZONE'
        43, // '(43) OPTION_ERO'
        44, // '(44) OPTION_LQ_QUERY'
        45, // '(45) OPTION_CLIENT_DATA'
        46, // '(46) OPTION_CLT_TIME'
        47, // '(47) OPTION_LQ_RELAY_DATA'
        48, // '(48) OPTION_LQ_CLIENT_LINK'
        // 49, // '(49) OPTION_MIP6_HNIDF'
        // 50, // '(50) OPTION_MIP6_VDINF'
        51, // '(51) OPTION_V6_LOST'
        52, // '(52) OPTION_CAPWAP_AC_V6'
        53, // '(53) OPTION_RELAY_ID'
        // 54, // '(54) OPTION-IPv6_Address-MoS'
        // 55, // '(55) OPTION-IPv6_FQDN-MoS'
        // 56, // '(56) OPTION_NTP_SERVER'
        57, // '(57) OPTION_V6_ACCESS_DOMAIN'
        58, // '(58) OPTION_SIP_UA_CS_LIST'
        59, // '(59) OPT_BOOTFILE_URL'
        60, // '(60) OPT_BOOTFILE_PARAM'
        61, // '(61) OPTION_CLIENT_ARCH_TYPE'
        62, // '(62) OPTION_NII'
        // 63, // '(63) OPTION_GEOLOCATION'
        64, // '(64) OPTION_AFTR_NAME'
        65, // '(65) OPTION_ERP_LOCAL_DOMAIN_NAME'
        66, // '(66) OPTION_RSOO'
        67, // '(67) OPTION_PD_EXCLUDE'
        // 68, // '(68) OPTION_VSS'
        // 69, // '(69) OPTION_MIP6_IDINF'
        // 70, // '(70) OPTION_MIP6_UDINF'
        // 71, // '(71) OPTION_MIP6_HNP'
        // 72, // '(72) OPTION_MIP6_HAA'
        // 73, // '(73) OPTION_MIP6_HAF'
        74, // '(74) OPTION_RDNSS_SELECTION'
        // 75, // '(75) OPTION_KRB_PRINCIPAL_NAME'
        // 76, // '(76) OPTION_KRB_REALM_NAME'
        // 77, // '(77) OPTION_KRB_DEFAULT_REALM_NAME'
        // 78, // '(78) OPTION_KRB_KDC'
        79, // '(79) OPTION_CLIENT_LINKLAYER_ADDR'
        80, // '(80) OPTION_LINK_ADDRESS'
        // 81, // '(81) OPTION_RADIUS'
        82, // '(82) OPTION_SOL_MAX_RT'
        83, // '(83) OPTION_INF_MAX_RT'
        // 84, // '(84) OPTION_ADDRSEL'
        // 85, // '(85) OPTION_ADDRSEL_TABLE'
        // 86, // '(86) OPTION_V6_PCP_SERVER'
        // 87, // '(87) OPTION_DHCPV4_MSG'
        88, // '(88) OPTION_DHCP4_O_DHCP6_SERVER'
        89, // '(89) OPTION_S46_RULE'
        90, // '(90) OPTION_S46_BR'
        91, // '(91) OPTION_S46_DMR'
        92, // '(92) OPTION_S46_V4V6BIND'
        93, // '(93) OPTION_S46_PORTPARAMS'
        94, // '(94) OPTION_S46_CONT_MAPE'
        95, // '(95) OPTION_S46_CONT_MAPT'
        96, // '(96) OPTION_S46_CONT_LW'
        // 97, // '(97) OPTION_4RD'
        // 98, // '(98) OPTION_4RD_MAP_RULE'
        // 99, // '(99) OPTION_4RD_NON_MAP_RULE'
        // 100, // '(100) OPTION_LQ_BASE_TIME'
        // 101, // '(101) OPTION_LQ_START_TIME'
        // 102, // '(102) OPTION_LQ_END_TIME'
        103, // '(103) DHCP Captive-Portal'
        // 104, // '(104) OPTION_MPL_PARAMETERS'
        // 105, // '(105) OPTION_ANI_ATT'
        // 106, // '(106) OPTION_ANI_NETWORK_NAME'
        // 107, // '(107) OPTION_ANI_AP_NAME'
        // 108, // '(108) OPTION_ANI_AP_BSSID'
        // 109, // '(109) OPTION_ANI_OPERATOR_ID'
        // 110, // '(110) OPTION_ANI_OPERATOR_REALM'
        // 111, // '(111) OPTION_S46_PRIORITY'
        // 112, // '(112) OPTION_MUD_URL_V6'
        // 113, // '(113) OPTION_V6_PREFIX64'
        // 114, // '(114) OPTION_F_BINDING_STATUS'
        // 115, // '(115) OPTION_F_CONNECT_FLAGS'
        // 116, // '(116) OPTION_F_DNS_REMOVAL_INFO'
        // 117, // '(117) OPTION_F_DNS_HOST_NAME'
        // 118, // '(118) OPTION_F_DNS_ZONE_NAME'
        // 119, // '(119) OPTION_F_DNS_FLAGS'
        // 120, // '(120) OPTION_F_EXPIRATION_TIME'
        // 121, // '(121) OPTION_F_MAX_UNACKED_BNDUPD'
        // 122, // '(122) OPTION_F_MCLT'
        // 123, // '(123) OPTION_F_PARTNER_LIFETIME'
        // 124, // '(124) OPTION_F_PARTNER_LIFETIME_SENT'
        // 125, // '(125) OPTION_F_PARTNER_DOWN_TIME'
        // 126, // '(126) OPTION_F_PARTNER_RAW_CLT_TIME'
        // 127, // '(127) OPTION_F_PROTOCOL_VERSION'
        // 128, // '(128) OPTION_F_KEEPALIVE_TIME'
        // 129, // '(129) OPTION_F_RECONFIGURE_DATA'
        // 130, // '(130) OPTION_F_RELATIONSHIP_NAME'
        // 131, // '(131) OPTION_F_SERVER_FLAGS'
        // 132, // '(132) OPTION_F_SERVER_STATE'
        // 133, // '(133) OPTION_F_START_TIME_OF_STATE'
        // 134, // '(134) OPTION_F_STATE_EXPIRATION_TIME'
        // 135, // '(135) OPTION_RELAY_PORT'
        // 136, // '(136) OPTION_V6_SZTP_REDIRECT'
        // 137, // '(137) OPTION_S46_BIND_IPV6_PREFIX'
        // 138, // '(138) OPTION_IA_LL'
        // 139, // '(139) OPTION_LLADDR'
        // 140, // '(140) OPTION_SLAP_QUAD'
        // 141, // '(141) OPTION_V6_DOTS_RI'
        // 142, // '(142) OPTION_V6_DOTS_ADDRESS'
        143, // '(143) OPTION-IPv6_Address-ANDSF'
    ]

    /**
     * Defines a list of configurable standard DHCPv6 options.
     *
     * Commented out options are not configurable by a user.
     */
    private _dhcpv6Options: DhcpOptionListItem[]

    /**
     * Indexes the standard DHCPv6 options by option code for faster lookup.
     */
    private _dhcpv6OptionsByCode: Map<number, DhcpOptionListItem>

    /**
     * Converts DHCP option definition to the list item.
     */
    private static convertToListItem(def: DhcpOptionDef): DhcpOptionListItem {
        return {
            label: DhcpOptionsService.getOptionLabel(def),
            value: def.code
        }
    }

    /**
     * Converts DHCP option definitions to the list items.
     */
    private static convertToListItems(defs: DhcpOptionDef[]): DhcpOptionListItem[] {
        return defs.map(DhcpOptionsService.convertToListItem)
    }

    /**
     * Constructor.
     *
     * Creates indexes of the options by the option codes.
     */
    constructor() {
        this._dhcpv4Options = stdDhcpv4OptionDefs
            .filter((def) => this.configurableDHCPv4OptionCodes.includes(def.code))
            .map(DhcpOptionsService.convertToListItem)
        this._dhcpv6Options = stdDhcpv6OptionDefs
            .filter((def) => this.configurableDHCPv6OptionCodes.includes(def.code))
            .map(DhcpOptionsService.convertToListItem)

        this._dhcpv4OptionsByCode = new Map(this._dhcpv4Options.map((o) => [o.value, o]))
        this._dhcpv6OptionsByCode = new Map(this._dhcpv6Options.map((o) => [o.value, o]))
    }

    /**
     * Returns the custom DHCP option definitions for a specific daemon.
     * 
     * @param daemonId daemon ID.
     */
    private async getCustomDhcpOptionDefinitions(daemonId: number): Promise<DhcpOptionDef[]> {
        throw new Error('Not implemented')
    }

    /**
     * Returns configurable standard DHCPv4 options.
     *
     * Returned list can be used to initialize dropdown list of options in a form.
     */
    private getStandardDhcpv4Options(): DhcpOptionListItem[] {
        return this._dhcpv4Options
    }

    /**
     * Returns configurable custom options for a given daemon.
     * 
     * Returned list can be used to initialize dropdown list of options in a form.
     * 
     * @param daemonId daemon ID.
     */
    private async getCustomDhcpOptions(daemonId: number): Promise<DhcpOptionListItem[]> {
        const defs = await this.getCustomDhcpOptionDefinitions(daemonId)
        return DhcpOptionsService.convertToListItems(defs)
    }

    /**
     * Returns configurable standard and custom DHCPv4 options.
     * 
     * Returned list can be used to initialize dropdown list of options in a form.
     * 
     * @param daemonId daemon ID.
     */
    async getDhcpv4Options(daemonId: number): Promise<DhcpOptionListItem[]> {
        const customDefs = await this.getCustomDhcpOptions(daemonId)
        return this.getStandardDhcpv4Options().concat(customDefs)
    }

    /**
     * Returns configurable standard DHCPv6 options.
     *
     * Returned list can be used to initialize dropdown list of options in a form.
     */
    private getStandardDhcpv6Options(): DhcpOptionListItem[] {
        return this._dhcpv6Options
    }

    /**
     * Returns configurable standard and custom DHCPv6 options.
     *
     * Returned list can be used to initialize dropdown list of options in a form.
     */
    async getDhcpv6Options(daemonId: number): Promise<DhcpOptionListItem[]> {
        const defs = await this.getCustomDhcpOptions(daemonId)
        return this.getStandardDhcpv6Options().concat(defs)
    }

    /**
     * Finds a specific DHCPv4 option by option code.
     *
     * @param code option code.
     * @returns option description or null if it is not found.
     */
    private findStandardDhcpv4Option(code: number): DhcpOptionListItem | null {
        return this._dhcpv4OptionsByCode.get(code)
    }

    /**
     * Finds a specific custom DHCP option by option code.
     *
     * @param daemonId daemon ID.
     * @param code option code.
     * @returns option description or null if it is not found.
     */
    private async findCustomDhcpOption(daemonId: number, code: number): Promise<DhcpOptionListItem | null> {
        const defs = await this.getCustomDhcpOptionDefinitions(daemonId)
        const def = defs.find((def) => def.code === code)
        return def ? DhcpOptionsService.convertToListItem(def) : null
    }

    /**
     * Finds a specific (standard or custom) DHCPv4 option by option code.
     *
     * @param code option code.
     * @returns option description or null if it is not found.
     */
    async findDhcpv4Option(daemonId: number, code: number): Promise<DhcpOptionListItem | null> {
        return this.findStandardDhcpv4Option(code) ?? this.findCustomDhcpOption(daemonId, code)
    }

    /**
     * Finds a specific DHCPv6 option by option code.
     *
     * @param code option code.
     * @returns option description or null if it is not found.
     */
    private findStandardDhcpv6Option(code: number): DhcpOptionListItem | null {
        return this._dhcpv6OptionsByCode.get(code)
    }

    /**
     * Finds a specific (standard or custom) DHCPv6 option by option code.
     *
     * @param code option code.
     * @returns option description or null if it is not found.
     */
    async findDhcpv6Option(daemonId: number, code: number): Promise<DhcpOptionListItem | null> {
        return this.findStandardDhcpv6Option(code) ?? this.findCustomDhcpOption(daemonId, code)
    }

    /**
     * Finds a standard DHCPv4 option definition by the code and space.
     *
     * @param code option code.
     * @param space option space.
     * @returns DHCPv4 option definition or null, if not found.
     */
    private findStandardDhcpv4OptionDef(code: number, space: string | null): DhcpOptionDef | null {
        return stdDhcpv4OptionDefs.find((def) => def.code === code && def.space === (space ?? 'dhcp4'))
    }

    /**
     * Finds a custom DHCPv4 option definition by the code and space.
     *
     * @param daemonId daemon ID.
     * @param code option code.
     * @param space option space.
     * @returns DHCPv4 option definition or null, if not found.
     */
    private async findCustomDhcpv4OptionDef(daemonId: number, code: number, space: string | null): Promise<DhcpOptionDef | null> {
        const defs = await this.getCustomDhcpOptionDefinitions(daemonId)
        return defs.find((def) => def.code === code && def.space === (space ?? 'dhcp4'))
    }

    /**
     * Finds a (standard or custom) DHCPv4 option definition by the code and space.
     * 
     * @param daemonId daemon ID.
     * @param code option code.
     * @param space option space.
     * @returns DHCPv4 option definition or null, if not found.
     */
    async findDhcpv4OptionDef(daemonId: number, code: number, space: string | null): Promise<DhcpOptionDef | null> {
        return this.findStandardDhcpv4OptionDef(code, space) ?? this.findCustomDhcpv4OptionDef(daemonId, code, space)
    }

    /**
     * Finds a standard DHCPv6 option definition by the code and space.
     *
     * @param code option code.
     * @param space option space.
     * @returns DHCPv6 option definition or null, if not found.
     */
    private findStandardDhcpv6OptionDef(code: number, space: string | null): DhcpOptionDef | null {
        return stdDhcpv6OptionDefs.find((def) => def.code === code && def.space === (space ?? 'dhcp6'))
    }

    /**
     * Finds a custom DHCPv6 option definition by the code and space.
     *
     * @param daemonId daemon ID.
     * @param code option code.
     * @param space option space.
     * @returns DHCPv6 option definition or null, if not found.
     */
    private async findCustomDhcpv6OptionDef(daemonId: number, code: number, space: string | null): Promise<DhcpOptionDef | null> {
        const defs = await this.getCustomDhcpOptionDefinitions(daemonId)
        return defs.find((def) => def.code === code && def.space === (space ?? 'dhcp6'))
    }

    /**
     * Finds a (standard or custom) DHCPv6 option definition by the code and space.
     * 
     * @param daemonId daemon ID.
     * @param code option code.
     * @param space option space.
     * @returns DHCPv6 option definition or null, if not found.
     */
    async findDhcpv6OptionDef(daemonId: number, code: number, space: string | null): Promise<DhcpOptionDef | null> {
        return this.findStandardDhcpv6OptionDef(code, space) ?? this.findCustomDhcpv6OptionDef(daemonId, code, space)
    }

    /**
     * Finds all standard DHCPv4 option definitions in option space.
     *
     * If the option space is null, the top-level dhcp4 option space is assumed.
     *
     * @param space option space name.
     * @returns An array of option definitions in the option space.
     */
    private  findStandardDhcpv4OptionDefsBySpace(space: string | null): DhcpOptionDef[] {
        return stdDhcpv4OptionDefs.filter((def) => def.space === (space ?? 'dhcp4'))
    }

    /**
     * Finds all custom DHCPv4 option definitions in option space.
     *
     * If the option space is null, the top-level dhcp4 option space is assumed.
     *
     * @param daemonId daemon ID.
     * @param space option space name.
     * @returns An array of option definitions in the option space.
     */
    private async findCustomDhcpv4OptionDefsBySpace(daemonId: number, space: string | null): Promise<DhcpOptionDef[]> {
        const defs = await this.getCustomDhcpOptionDefinitions(daemonId)
        return defs.filter((def) => def.space === (space ?? 'dhcp4'))
    }

    /**
     * Finds all (standard or custom) DHCPv4 option definitions in option space.
     *
     * If the option space is null, the top-level dhcp6 option space is assumed.
     *
     * @param daemonId daemon ID.
     * @param space option space name.
     * @returns An array of option definitions in the option space.
     */
    async findDhcpv4OptionDefsBySpace(daemonId: number, space: string | null): Promise<DhcpOptionDef[]> {
        const customDefs = await this.findCustomDhcpv4OptionDefsBySpace(daemonId, space)
        return this.findStandardDhcpv4OptionDefsBySpace(space).concat(customDefs)
    }

    /**
     * Finds all standard DHCPv6 option definitions in option space.
     *
     * If the option space is null, the top-level dhcp6 option space is assumed.
     *
     * @param space option space name.
     * @returns An array of option definitions in the option space.
     */
    private findStandardDhcpv6OptionDefsBySpace(space: string | null): DhcpOptionDef[] {
        return stdDhcpv6OptionDefs.filter((def) => def.space === (space ?? 'dhcp6'))
    }

    /**
     * Finds all custom DHCPv6 option definitions in option space.
     *
     * If the option space is null, the top-level dhcp6 option space is assumed.
     *
     * @param daemonId daemon ID.
     * @param space option space name.
     * @returns An array of option definitions in the option space.
     */
    private async findCustomDhcpv6OptionDefsBySpace(daemonId: number, space: string | null): Promise<DhcpOptionDef[]> {
        const defs = await this.getCustomDhcpOptionDefinitions(daemonId)
        return defs.filter((def) => def.space === (space ?? 'dhcp6'))
    }


    /**
     * Finds all (standard or custom) DHCPv6 option definitions in option space.
     *
     * If the option space is null, the top-level dhcp6 option space is assumed.
     *
     * @param daemonId daemon ID.
     * @param space option space name.
     * @returns An array of option definitions in the option space.
     */
    async findDhcpv6OptionDefsBySpace(daemonId: number, space: string | null): Promise<DhcpOptionDef[]> {
        const customDefs = await this.findCustomDhcpv6OptionDefsBySpace(daemonId, space)
        return this.findStandardDhcpv6OptionDefsBySpace(space).concat(customDefs)
    }

    /**
     * Constructs a conventional label for a DHCP option definition.
     */
    private static getOptionLabel(def: DhcpOptionDef): string {
        return `(${def.code}) ${def.name.replace('-', '_').toUpperCase()}`
    }
}
