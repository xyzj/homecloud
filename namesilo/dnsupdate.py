#!/bin/env python

import sys
 
if sys.version_info.major == 2:
    # Python2
    from urllib import urlencode
    from urllib import quote
    from urlparse import urlparse
    import urllib2 as request
else:
    # Python3
    from urllib.parse import urlencode  
    from urllib.parse import quote
    from urllib.parse import urlparse
    import urllib.request as request

from xml.etree.ElementTree import parse as xmlparse

NAMESILO_KEY = "f59e74d5e3f373b9e332e9b"
NAMESILO_RRID = ""

URL_MYIP = "https://wgq.shwlst.com:40001/"
URL_LISTDNS = "https://www.namesilo.com/api/dnsListRecords?version=1&type=xml&key={0}&domain=xyzjdays.xyz"
URL_UPDATENDS = "https://www.namesilo.com/api/dnsUpdateRecord?version=1&type=xml&key={0}&domain=xyzjdays.xyz&rrid={1}&rrhost=hpc&rrvalue={2}"

if __name__=="__main__":
    hdr = {'User-Agent': 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.64 Safari/537.11',
    'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
    'Accept-Charset': 'ISO-8859-1,utf-8;q=0.7,*;q=0.3',
    'Accept-Encoding': 'none',
    'Accept-Language': 'en-US,en;q=0.8',
    'Connection': 'keep-alive'}
            
    # read new ip
    try:
        resp = request.urlopen(URL_MYIP,timeout=10)
    except:
        print("can not get the new ip, stop the script.")
        sys.exit(1)
    newip= resp.read()
    print("current ip is: "+newip)

    # read old ip
    with open(".ipcache", "r") as f:
        oldip=f.read(15).strip()
        f.close()
        print("cached ip is: "+oldip)

    if newip != oldip: # ip not same, update dns
        # view rrid
        req = request.Request(URL_LISTDNS.format(NAMESILO_KEY),headers=hdr)
        # print(URL_LISTDNS.format(NAMESILO_KEY))
        try:
            resp = request.urlopen(req)
        except Exception as ex:
            print(ex)
            sys.exit(1)

        doc = xmlparse(resp)
        allrecords = doc.findall("reply/resource_record")
        for record in allrecords:
            if record.find("host").text=="hpc.xyzjdays.xyz":
                NAMESILO_RRID=record.find("record_id").text
                print("record id is: "+NAMESILO_RRID)
                break
        
        if NAMESILO_RRID!="":
            # update dns
            req = request.Request(URL_UPDATENDS.format(NAMESILO_KEY,NAMESILO_RRID,newip),headers=hdr)
            try:
                resp = request.urlopen(req)
            except Exception as ex:
                print(ex)
                sys.exit(1)
            doc = xmlparse(resp)
            print(doc.findtext("reply/code"),doc.findtext("reply/detail"))
            # success update .ipcache file
            if doc.findtext("reply/code") == "300":
                with open(".ipcache","w+") as f:
                    f.write(newip)
                    f.close()