import { TestBed } from '@angular/core/testing'

import { DhcpOptionsService } from './dhcp-options.service'
import stdDhcpv4OptionDefs from './std-dhcpv4-option-defs.json'
import stdDhcpv6OptionDefs from './std-dhcpv6-option-defs.json'
import { DHCPOptionDefinitions, DHCPService } from './backend'
import { of } from 'rxjs'
import { HttpResponse } from '@angular/common/http'
import { HttpClientTestingModule } from '@angular/common/http/testing'

describe('DhcpOptionsService', async () => {
    let service: DhcpOptionsService

    beforeEach(() => {
        TestBed.configureTestingModule({ imports: [HttpClientTestingModule] })
        service = TestBed.inject(DhcpOptionsService)

        const dhcpService = TestBed.inject(DHCPService)
        spyOn(dhcpService, "getCustomOptionDefinitions").and.
            returnValue(of({
                total: 3,
                items: [
                    {
                        code: 1001,
                        name: "foo",
                        optionType: "uint8",
                        space: "dhcp4",
                    },
                    {
                        code: 1002,
                        name: "bar",
                        optionType: "uint16",
                        space: "dhcp4",
                        array: false,
                        recordTypes: ["uint16"]
                    },
                    {
                        code: 1003,
                        name: "baz",
                        optionType: "ipv4-address",
                        space: "zab",
                        array: true
                    }
                ]
            } as DHCPOptionDefinitions) as any) 
    })

    it('should be created', async () => {
        expect(service).toBeTruthy()
    })

    it('should return all configurable DHCPv4 options',async () => {
        const options = await service.getConfigurableDhcpv4OptionDefs(42)
        const listItems = service.convertToListItems(options)
        expect(listItems.length).toBe(98)

        // Validate one of them to make sure they are DHCPv4 options.
        const selectedItem = listItems.find((o) => o.value === 5)
        expect(selectedItem).toBeTruthy()
        expect(selectedItem.label).toBe('(5) Name Server')
    })

    it('should return all configurable DHCPv6 options', async () => {
        const options = await service.getConfigurableDhcpv6OptionDefs(42)
        const listItems = service.convertToListItems(options)
        expect(listItems.length).toBe(56)

        // Validate one of them to make sure they are DHCPv6 options.
        const selectedItem = listItems.find((o) => o.value === 23)
        expect(selectedItem).toBeTruthy()
        expect(selectedItem.label).toBe('(23) OPTION_DNS_SERVERS')
    })
})
