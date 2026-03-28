package decodo

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	defaultEndpoint              = "gate.decodo.com"
	defaultPort                  = 7000
	defaultStickyDurationMinutes = 10
)

// EndpointPreset describes a known Decodo endpoint with its rotating port and sticky port range.
type EndpointPreset struct {
	Host            string
	RotatingPort    int
	StickyPortRange PortRange
}

// wellKnownPresets contains all documented Decodo country endpoint configurations.
// Data sourced from https://help.decodo.com/docs/residential-proxy-endpoints-and-ports
var wellKnownPresets = map[string]EndpointPreset{
	"gate": {Host: "gate.decodo.com", RotatingPort: 10000, StickyPortRange: PortRange{Start: 10001, End: 49999}},
	"us":   {Host: "us.decodo.com", RotatingPort: 10000, StickyPortRange: PortRange{Start: 10001, End: 29999}},
	"eu":   {Host: "eu.decodo.com", RotatingPort: 10000, StickyPortRange: PortRange{Start: 10001, End: 29999}},
	"ae":   {Host: "ae.decodo.com", RotatingPort: 20000, StickyPortRange: PortRange{Start: 20001, End: 29999}},
	"my":   {Host: "my.decodo.com", RotatingPort: 30000, StickyPortRange: PortRange{Start: 30001, End: 39999}},
	"ph":   {Host: "ph.decodo.com", RotatingPort: 40000, StickyPortRange: PortRange{Start: 40001, End: 49999}},
	"in":   {Host: "in.decodo.com", RotatingPort: 10000, StickyPortRange: PortRange{Start: 10001, End: 19999}},
	"tw":   {Host: "tw.decodo.com", RotatingPort: 20000, StickyPortRange: PortRange{Start: 20001, End: 29999}},
	"jp":   {Host: "jp.decodo.com", RotatingPort: 30000, StickyPortRange: PortRange{Start: 30001, End: 39999}},
	"be":   {Host: "be.decodo.com", RotatingPort: 40000, StickyPortRange: PortRange{Start: 40001, End: 49999}},
	"es":   {Host: "es.decodo.com", RotatingPort: 10000, StickyPortRange: PortRange{Start: 10001, End: 19999}},
	"pt":   {Host: "pt.decodo.com", RotatingPort: 20000, StickyPortRange: PortRange{Start: 20001, End: 29999}},
	"gr":   {Host: "gr.decodo.com", RotatingPort: 30000, StickyPortRange: PortRange{Start: 30001, End: 39999}},
	"pe":   {Host: "pe.decodo.com", RotatingPort: 40000, StickyPortRange: PortRange{Start: 40001, End: 49999}},
	"ar":   {Host: "ar.decodo.com", RotatingPort: 10000, StickyPortRange: PortRange{Start: 10001, End: 19999}},
	"se":   {Host: "se.decodo.com", RotatingPort: 20000, StickyPortRange: PortRange{Start: 20001, End: 29999}},
	"az":   {Host: "az.decodo.com", RotatingPort: 30000, StickyPortRange: PortRange{Start: 30001, End: 39999}},
	"ua":   {Host: "ua.decodo.com", RotatingPort: 40000, StickyPortRange: PortRange{Start: 40001, End: 49999}},
	"hk":   {Host: "hk.decodo.com", RotatingPort: 10000, StickyPortRange: PortRange{Start: 10001, End: 19999}},
	"de":   {Host: "de.decodo.com", RotatingPort: 20000, StickyPortRange: PortRange{Start: 20001, End: 29999}},
	"ir":   {Host: "ir.decodo.com", RotatingPort: 30000, StickyPortRange: PortRange{Start: 30001, End: 39999}},
	"za":   {Host: "za.decodo.com", RotatingPort: 40000, StickyPortRange: PortRange{Start: 40001, End: 49999}},
	"kr":   {Host: "kr.decodo.com", RotatingPort: 10000, StickyPortRange: PortRange{Start: 10001, End: 19999}},
	"ec":   {Host: "ec.decodo.com", RotatingPort: 20000, StickyPortRange: PortRange{Start: 20001, End: 29999}},
	"cl":   {Host: "cl.decodo.com", RotatingPort: 30000, StickyPortRange: PortRange{Start: 30001, End: 39999}},
	"ru":   {Host: "ru.decodo.com", RotatingPort: 40000, StickyPortRange: PortRange{Start: 40001, End: 49999}},
	"id":   {Host: "id.decodo.com", RotatingPort: 10000, StickyPortRange: PortRange{Start: 10001, End: 19999}},
	"eg":   {Host: "eg.decodo.com", RotatingPort: 20000, StickyPortRange: PortRange{Start: 20001, End: 29999}},
	"cn":   {Host: "cn.decodo.com", RotatingPort: 30000, StickyPortRange: PortRange{Start: 30001, End: 39999}},
	"gb":   {Host: "gb.decodo.com", RotatingPort: 30000, StickyPortRange: PortRange{Start: 30001, End: 49999}},
	"nl":   {Host: "nl.decodo.com", RotatingPort: 10000, StickyPortRange: PortRange{Start: 10001, End: 19999}},
	"it":   {Host: "it.decodo.com", RotatingPort: 20000, StickyPortRange: PortRange{Start: 20001, End: 29999}},
	"au":   {Host: "au.decodo.com", RotatingPort: 30000, StickyPortRange: PortRange{Start: 30001, End: 39999}},
	"kz":   {Host: "kz.decodo.com", RotatingPort: 40000, StickyPortRange: PortRange{Start: 40001, End: 49999}},
	"sg":   {Host: "sg.decodo.com", RotatingPort: 10000, StickyPortRange: PortRange{Start: 10001, End: 19999}},
	"mx":   {Host: "mx.decodo.com", RotatingPort: 20000, StickyPortRange: PortRange{Start: 20001, End: 29999}},
	"th":   {Host: "th.decodo.com", RotatingPort: 30000, StickyPortRange: PortRange{Start: 30001, End: 39999}},
	"tr":   {Host: "tr.decodo.com", RotatingPort: 40000, StickyPortRange: PortRange{Start: 40001, End: 49999}},
	"br":   {Host: "br.decodo.com", RotatingPort: 10000, StickyPortRange: PortRange{Start: 10001, End: 19999}},
	"pl":   {Host: "pl.decodo.com", RotatingPort: 20000, StickyPortRange: PortRange{Start: 20001, End: 29999}},
	"co":   {Host: "co.decodo.com", RotatingPort: 30000, StickyPortRange: PortRange{Start: 30001, End: 39999}},
	"fr":   {Host: "fr.decodo.com", RotatingPort: 40000, StickyPortRange: PortRange{Start: 40001, End: 49999}},
	"pk":   {Host: "pk.decodo.com", RotatingPort: 10000, StickyPortRange: PortRange{Start: 10001, End: 19999}},
	"ca":   {Host: "ca.decodo.com", RotatingPort: 20000, StickyPortRange: PortRange{Start: 20001, End: 29999}},
	"il":   {Host: "il.decodo.com", RotatingPort: 30000, StickyPortRange: PortRange{Start: 30001, End: 39999}},
	"ma":   {Host: "ma.decodo.com", RotatingPort: 40000, StickyPortRange: PortRange{Start: 40001, End: 40999}},
	"mz":   {Host: "mz.decodo.com", RotatingPort: 41000, StickyPortRange: PortRange{Start: 41001, End: 41999}},
	"ng":   {Host: "ng.decodo.com", RotatingPort: 42000, StickyPortRange: PortRange{Start: 42001, End: 42999}},
	"gh":   {Host: "gh.decodo.com", RotatingPort: 43000, StickyPortRange: PortRange{Start: 43001, End: 43999}},
	"ci":   {Host: "ci.decodo.com", RotatingPort: 44000, StickyPortRange: PortRange{Start: 44001, End: 44999}},
	"ke":   {Host: "ke.decodo.com", RotatingPort: 45000, StickyPortRange: PortRange{Start: 45001, End: 45999}},
	"lr":   {Host: "lr.decodo.com", RotatingPort: 46000, StickyPortRange: PortRange{Start: 46001, End: 46999}},
	"mg":   {Host: "mg.decodo.com", RotatingPort: 47000, StickyPortRange: PortRange{Start: 47001, End: 47999}},
	"ml":   {Host: "ml.decodo.com", RotatingPort: 48000, StickyPortRange: PortRange{Start: 48001, End: 48999}},
	"mt":   {Host: "mt.decodo.com", RotatingPort: 49000, StickyPortRange: PortRange{Start: 49001, End: 49999}},
	"mc":   {Host: "mc.decodo.com", RotatingPort: 10000, StickyPortRange: PortRange{Start: 10001, End: 10999}},
	"md":   {Host: "md.decodo.com", RotatingPort: 11000, StickyPortRange: PortRange{Start: 11001, End: 11999}},
	"me":   {Host: "me.decodo.com", RotatingPort: 12000, StickyPortRange: PortRange{Start: 12001, End: 12999}},
	"no":   {Host: "no.decodo.com", RotatingPort: 13000, StickyPortRange: PortRange{Start: 13001, End: 13999}},
	"py":   {Host: "py.decodo.com", RotatingPort: 14000, StickyPortRange: PortRange{Start: 14001, End: 14999}},
	"uy":   {Host: "uy.decodo.com", RotatingPort: 15000, StickyPortRange: PortRange{Start: 15001, End: 15999}},
	"ve":   {Host: "ve.decodo.com", RotatingPort: 16000, StickyPortRange: PortRange{Start: 16001, End: 16999}},
	"dm":   {Host: "dm.decodo.com", RotatingPort: 17000, StickyPortRange: PortRange{Start: 17001, End: 17999}},
	"ht":   {Host: "ht.decodo.com", RotatingPort: 18000, StickyPortRange: PortRange{Start: 18001, End: 18999}},
	"hn":   {Host: "hn.decodo.com", RotatingPort: 19000, StickyPortRange: PortRange{Start: 19001, End: 19999}},
	"jm":   {Host: "jm.decodo.com", RotatingPort: 20000, StickyPortRange: PortRange{Start: 20001, End: 20999}},
	"aw":   {Host: "aw.smartproxy.com", RotatingPort: 21000, StickyPortRange: PortRange{Start: 21001, End: 21999}},
	"lv":   {Host: "lv.decodo.com", RotatingPort: 22000, StickyPortRange: PortRange{Start: 22001, End: 22999}},
	"li":   {Host: "li.decodo.com", RotatingPort: 23000, StickyPortRange: PortRange{Start: 23001, End: 23999}},
	"lt":   {Host: "lt.decodo.com", RotatingPort: 24000, StickyPortRange: PortRange{Start: 24001, End: 24999}},
	"lu":   {Host: "lu.decodo.com", RotatingPort: 25000, StickyPortRange: PortRange{Start: 25001, End: 25999}},
	"jo":   {Host: "jo.decodo.com", RotatingPort: 26000, StickyPortRange: PortRange{Start: 26001, End: 26999}},
	"lb":   {Host: "lb.decodo.com", RotatingPort: 27000, StickyPortRange: PortRange{Start: 27001, End: 27999}},
	"mv":   {Host: "mv.decodo.com", RotatingPort: 28000, StickyPortRange: PortRange{Start: 28001, End: 28999}},
	"mn":   {Host: "mn.decodo.com", RotatingPort: 29000, StickyPortRange: PortRange{Start: 29001, End: 29999}},
	"om":   {Host: "om.decodo.com", RotatingPort: 30000, StickyPortRange: PortRange{Start: 30001, End: 30999}},
	"sd":   {Host: "sd.decodo.com", RotatingPort: 31000, StickyPortRange: PortRange{Start: 31001, End: 31999}},
	"tg":   {Host: "tg.decodo.com", RotatingPort: 32000, StickyPortRange: PortRange{Start: 32001, End: 32999}},
	"tn":   {Host: "tn.decodo.com", RotatingPort: 33000, StickyPortRange: PortRange{Start: 33001, End: 33999}},
	"ug":   {Host: "ug.decodo.com", RotatingPort: 34000, StickyPortRange: PortRange{Start: 34001, End: 34999}},
	"zm":   {Host: "zm.decodo.com", RotatingPort: 35000, StickyPortRange: PortRange{Start: 35001, End: 35999}},
	"af":   {Host: "af.decodo.com", RotatingPort: 36000, StickyPortRange: PortRange{Start: 36001, End: 36999}},
	"bh":   {Host: "bh.decodo.com", RotatingPort: 37000, StickyPortRange: PortRange{Start: 37001, End: 37999}},
	"fj":   {Host: "fj.decodo.com", RotatingPort: 38000, StickyPortRange: PortRange{Start: 38001, End: 38999}},
	"nz":   {Host: "nz.decodo.com", RotatingPort: 39000, StickyPortRange: PortRange{Start: 39001, End: 39999}},
	"bo":   {Host: "bo.decodo.com", RotatingPort: 40000, StickyPortRange: PortRange{Start: 40001, End: 40999}},
	"bd":   {Host: "bd.decodo.com", RotatingPort: 41001, StickyPortRange: PortRange{Start: 41001, End: 41099}},
	"am":   {Host: "am.decodo.com", RotatingPort: 42000, StickyPortRange: PortRange{Start: 42001, End: 42999}},
	"ge":   {Host: "ge.decodo.com", RotatingPort: 43000, StickyPortRange: PortRange{Start: 43001, End: 43999}},
	"iq":   {Host: "iq.decodo.com", RotatingPort: 44000, StickyPortRange: PortRange{Start: 44001, End: 44999}},
	"bt":   {Host: "bt.decodo.com", RotatingPort: 45000, StickyPortRange: PortRange{Start: 45001, End: 45999}},
	"mm":   {Host: "mm.decodo.com", RotatingPort: 46000, StickyPortRange: PortRange{Start: 46001, End: 46999}},
	"kh":   {Host: "kh.decodo.com", RotatingPort: 47000, StickyPortRange: PortRange{Start: 47001, End: 47999}},
	"cy":   {Host: "cy.decodo.com", RotatingPort: 48000, StickyPortRange: PortRange{Start: 48001, End: 48999}},
	"sn":   {Host: "sn.decodo.com", RotatingPort: 49000, StickyPortRange: PortRange{Start: 49001, End: 49999}},
	"sc":   {Host: "sc.decodo.com", RotatingPort: 10000, StickyPortRange: PortRange{Start: 10001, End: 10999}},
	"zw":   {Host: "zw.decodo.com", RotatingPort: 11000, StickyPortRange: PortRange{Start: 11001, End: 11999}},
	"ss":   {Host: "ss.decodo.com", RotatingPort: 12000, StickyPortRange: PortRange{Start: 12001, End: 12999}},
	"ro":   {Host: "ro.decodo.com", RotatingPort: 13000, StickyPortRange: PortRange{Start: 13001, End: 13999}},
	"rs":   {Host: "rs.decodo.com", RotatingPort: 14000, StickyPortRange: PortRange{Start: 14001, End: 14999}},
	"sk":   {Host: "sk.decodo.com", RotatingPort: 15000, StickyPortRange: PortRange{Start: 15001, End: 15999}},
	"si":   {Host: "si.decodo.com", RotatingPort: 16000, StickyPortRange: PortRange{Start: 16001, End: 16999}},
	"bs":   {Host: "bs.decodo.com", RotatingPort: 17000, StickyPortRange: PortRange{Start: 17001, End: 17999}},
	"bz":   {Host: "bz.decodo.com", RotatingPort: 18000, StickyPortRange: PortRange{Start: 18001, End: 18999}},
	"vg":   {Host: "vg.decodo.com", RotatingPort: 19000, StickyPortRange: PortRange{Start: 19001, End: 19999}},
	"pa":   {Host: "pa.decodo.com", RotatingPort: 20000, StickyPortRange: PortRange{Start: 20001, End: 20999}},
	"pr":   {Host: "pr.decodo.com", RotatingPort: 21000, StickyPortRange: PortRange{Start: 21001, End: 21999}},
	"tt":   {Host: "tt.decodo.com", RotatingPort: 22000, StickyPortRange: PortRange{Start: 22001, End: 22999}},
	"is":   {Host: "is.decodo.com", RotatingPort: 23000, StickyPortRange: PortRange{Start: 23001, End: 23999}},
	"ie":   {Host: "ie.decodo.com", RotatingPort: 24000, StickyPortRange: PortRange{Start: 24001, End: 24999}},
	"cz":   {Host: "cz.decodo.com", RotatingPort: 26000, StickyPortRange: PortRange{Start: 26001, End: 26999}},
	"dk":   {Host: "dk.decodo.com", RotatingPort: 27000, StickyPortRange: PortRange{Start: 27001, End: 27999}},
	"ee":   {Host: "ee.decodo.com", RotatingPort: 28000, StickyPortRange: PortRange{Start: 28001, End: 28999}},
	"ch":   {Host: "ch.decodo.com", RotatingPort: 29000, StickyPortRange: PortRange{Start: 29001, End: 29999}},
	"mk":   {Host: "mk.decodo.com", RotatingPort: 30000, StickyPortRange: PortRange{Start: 30001, End: 30999}},
	"cr":   {Host: "cr.decodo.com", RotatingPort: 31000, StickyPortRange: PortRange{Start: 31001, End: 31999}},
	"cu":   {Host: "cu.decodo.com", RotatingPort: 32000, StickyPortRange: PortRange{Start: 32001, End: 32999}},
	"al":   {Host: "al.decodo.com", RotatingPort: 33000, StickyPortRange: PortRange{Start: 33001, End: 33999}},
	"ad":   {Host: "ad.decodo.com", RotatingPort: 34000, StickyPortRange: PortRange{Start: 34001, End: 34999}},
	"at":   {Host: "at.decodo.com", RotatingPort: 35000, StickyPortRange: PortRange{Start: 35001, End: 35999}},
	"ba":   {Host: "ba.decodo.com", RotatingPort: 37000, StickyPortRange: PortRange{Start: 37001, End: 37999}},
	"bg":   {Host: "bg.decodo.com", RotatingPort: 38000, StickyPortRange: PortRange{Start: 38001, End: 38999}},
	"by":   {Host: "by.decodo.com", RotatingPort: 39000, StickyPortRange: PortRange{Start: 39001, End: 39999}},
	"hr":   {Host: "hr.decodo.com", RotatingPort: 40000, StickyPortRange: PortRange{Start: 40001, End: 40999}},
	"fi":   {Host: "fi.decodo.com", RotatingPort: 41000, StickyPortRange: PortRange{Start: 41001, End: 41099}},
	"hu":   {Host: "hu.decodo.com", RotatingPort: 43000, StickyPortRange: PortRange{Start: 43001, End: 43999}},
	"qa":   {Host: "qa.decodo.com", RotatingPort: 44000, StickyPortRange: PortRange{Start: 44001, End: 44999}},
	"sa":   {Host: "sa.decodo.com", RotatingPort: 45000, StickyPortRange: PortRange{Start: 45001, End: 45999}},
	"vn":   {Host: "vn.decodo.com", RotatingPort: 46000, StickyPortRange: PortRange{Start: 46001, End: 46999}},
	"tm":   {Host: "tm.decodo.com", RotatingPort: 47000, StickyPortRange: PortRange{Start: 47001, End: 47999}},
	"uz":   {Host: "uz.decodo.com", RotatingPort: 48000, StickyPortRange: PortRange{Start: 48001, End: 48999}},
	"ye":   {Host: "ye.decodo.com", RotatingPort: 49000, StickyPortRange: PortRange{Start: 49001, End: 49999}},
	"cf":   {Host: "cf.decodo.com", RotatingPort: 10000, StickyPortRange: PortRange{Start: 10001, End: 10999}},
	"td":   {Host: "td.decodo.com", RotatingPort: 11000, StickyPortRange: PortRange{Start: 11001, End: 11999}},
	"bj":   {Host: "bj.decodo.com", RotatingPort: 12000, StickyPortRange: PortRange{Start: 12001, End: 12999}},
	"et":   {Host: "et.decodo.com", RotatingPort: 13000, StickyPortRange: PortRange{Start: 13001, End: 13999}},
	"dj":   {Host: "dj.decodo.com", RotatingPort: 14000, StickyPortRange: PortRange{Start: 14001, End: 14999}},
	"gm":   {Host: "gm.decodo.com", RotatingPort: 15000, StickyPortRange: PortRange{Start: 15001, End: 15999}},
	"mr":   {Host: "mr.decodo.com", RotatingPort: 16000, StickyPortRange: PortRange{Start: 16001, End: 16999}},
	"mu":   {Host: "mu.decodo.com", RotatingPort: 17000, StickyPortRange: PortRange{Start: 17001, End: 17999}},
	"ao":   {Host: "ao.decodo.com", RotatingPort: 18000, StickyPortRange: PortRange{Start: 18001, End: 18999}},
	"cm":   {Host: "cm.decodo.com", RotatingPort: 19000, StickyPortRange: PortRange{Start: 19001, End: 19999}},
	"sy":   {Host: "sy.decodo.com", RotatingPort: 20000, StickyPortRange: PortRange{Start: 20001, End: 20999}},
}

// cityPresets contains all documented city-level endpoints with their specific ports.
// Data sourced from https://help.decodo.com/docs/residential-proxy-endpoints-and-ports
var cityPresets = map[string]EndpointPreset{
	"new_york":     {Host: "city.decodo.com", RotatingPort: 21000, StickyPortRange: PortRange{Start: 21001, End: 21049}},
	"los_angeles":  {Host: "city.decodo.com", RotatingPort: 21050, StickyPortRange: PortRange{Start: 21051, End: 21099}},
	"chicago":      {Host: "city.decodo.com", RotatingPort: 21000, StickyPortRange: PortRange{Start: 21101, End: 21149}},
	"houston":      {Host: "city.decodo.com", RotatingPort: 21150, StickyPortRange: PortRange{Start: 21151, End: 21199}},
	"miami":        {Host: "city.decodo.com", RotatingPort: 21200, StickyPortRange: PortRange{Start: 21201, End: 21249}},
	"london":       {Host: "city.decodo.com", RotatingPort: 21250, StickyPortRange: PortRange{Start: 21251, End: 21299}},
	"berlin":       {Host: "city.decodo.com", RotatingPort: 21300, StickyPortRange: PortRange{Start: 21301, End: 21349}},
	"moscow":       {Host: "city.decodo.com", RotatingPort: 21350, StickyPortRange: PortRange{Start: 21351, End: 21399}},
}

// statePresets contains all documented US state-level endpoints with their specific ports.
// Data sourced from https://help.decodo.com/docs/residential-proxy-endpoints-and-ports
var statePresets = map[string]EndpointPreset{
	"alabama":        {Host: "state.decodo.com", RotatingPort: 17000, StickyPortRange: PortRange{Start: 17001, End: 17099}},
	"alaska":         {Host: "state.decodo.com", RotatingPort: 17100, StickyPortRange: PortRange{Start: 17101, End: 17199}},
	"arizona":        {Host: "state.decodo.com", RotatingPort: 17200, StickyPortRange: PortRange{Start: 17201, End: 17299}},
	"arkansas":       {Host: "state.decodo.com", RotatingPort: 17300, StickyPortRange: PortRange{Start: 17301, End: 17399}},
	"california":     {Host: "state.decodo.com", RotatingPort: 10000, StickyPortRange: PortRange{Start: 10001, End: 10999}},
	"colorado":       {Host: "state.decodo.com", RotatingPort: 17400, StickyPortRange: PortRange{Start: 17401, End: 17499}},
	"connecticut":    {Host: "state.decodo.com", RotatingPort: 17500, StickyPortRange: PortRange{Start: 17501, End: 17599}},
	"delaware":       {Host: "state.decodo.com", RotatingPort: 17600, StickyPortRange: PortRange{Start: 17601, End: 17699}},
	"florida":        {Host: "state.decodo.com", RotatingPort: 11000, StickyPortRange: PortRange{Start: 11001, End: 11999}},
	"georgia":        {Host: "state.decodo.com", RotatingPort: 17700, StickyPortRange: PortRange{Start: 17701, End: 17799}},
	"hawaii":         {Host: "state.decodo.com", RotatingPort: 17800, StickyPortRange: PortRange{Start: 17801, End: 17899}},
	"idaho":          {Host: "state.decodo.com", RotatingPort: 17900, StickyPortRange: PortRange{Start: 17901, End: 17999}},
	"illinois":       {Host: "state.decodo.com", RotatingPort: 12000, StickyPortRange: PortRange{Start: 12001, End: 12999}},
	"indiana":         {Host: "state.decodo.com", RotatingPort: 18000, StickyPortRange: PortRange{Start: 18001, End: 18099}},
	"iowa":           {Host: "state.decodo.com", RotatingPort: 18100, StickyPortRange: PortRange{Start: 18101, End: 18199}},
	"kansas":         {Host: "state.decodo.com", RotatingPort: 18200, StickyPortRange: PortRange{Start: 18201, End: 18299}},
	"kentucky":       {Host: "state.decodo.com", RotatingPort: 18300, StickyPortRange: PortRange{Start: 18301, End: 18399}},
	"louisiana":      {Host: "state.decodo.com", RotatingPort: 18400, StickyPortRange: PortRange{Start: 18401, End: 18499}},
	"maine":          {Host: "state.decodo.com", RotatingPort: 18500, StickyPortRange: PortRange{Start: 18501, End: 18599}},
	"maryland":       {Host: "state.decodo.com", RotatingPort: 18600, StickyPortRange: PortRange{Start: 18601, End: 18699}},
	"massachusetts":   {Host: "state.decodo.com", RotatingPort: 18700, StickyPortRange: PortRange{Start: 18701, End: 18799}},
	"michigan":       {Host: "state.decodo.com", RotatingPort: 18800, StickyPortRange: PortRange{Start: 18801, End: 18899}},
	"minnesota":      {Host: "state.decodo.com", RotatingPort: 18900, StickyPortRange: PortRange{Start: 18901, End: 18999}},
	"mississippi":    {Host: "state.decodo.com", RotatingPort: 19000, StickyPortRange: PortRange{Start: 19001, End: 19099}},
	"missouri":       {Host: "state.decodo.com", RotatingPort: 19100, StickyPortRange: PortRange{Start: 19101, End: 19199}},
	"montana":        {Host: "state.decodo.com", RotatingPort: 19200, StickyPortRange: PortRange{Start: 19201, End: 19299}},
	"nebraska":       {Host: "state.decodo.com", RotatingPort: 19300, StickyPortRange: PortRange{Start: 19301, End: 19399}},
	"nevada":         {Host: "state.decodo.com", RotatingPort: 19400, StickyPortRange: PortRange{Start: 19401, End: 19499}},
	"new_hampshire":  {Host: "state.decodo.com", RotatingPort: 19500, StickyPortRange: PortRange{Start: 19501, End: 19599}},
	"new_jersey":     {Host: "state.decodo.com", RotatingPort: 19600, StickyPortRange: PortRange{Start: 19601, End: 19699}},
	"new_mexico":     {Host: "state.decodo.com", RotatingPort: 19700, StickyPortRange: PortRange{Start: 19701, End: 19799}},
	"new_york":       {Host: "state.decodo.com", RotatingPort: 13000, StickyPortRange: PortRange{Start: 13001, End: 13999}},
	"north_carolina": {Host: "state.decodo.com", RotatingPort: 19800, StickyPortRange: PortRange{Start: 19801, End: 19899}},
	"north_dakota":   {Host: "state.decodo.com", RotatingPort: 19900, StickyPortRange: PortRange{Start: 19901, End: 19999}},
	"ohio":           {Host: "state.decodo.com", RotatingPort: 20000, StickyPortRange: PortRange{Start: 20001, End: 20099}},
	"oklahoma":       {Host: "state.decodo.com", RotatingPort: 20100, StickyPortRange: PortRange{Start: 20101, End: 20199}},
	"oregon":         {Host: "state.decodo.com", RotatingPort: 20200, StickyPortRange: PortRange{Start: 20201, End: 20299}},
	"pennsylvania":   {Host: "state.decodo.com", RotatingPort: 20300, StickyPortRange: PortRange{Start: 20301, End: 20399}},
	"rhode_island":   {Host: "state.decodo.com", RotatingPort: 20400, StickyPortRange: PortRange{Start: 20401, End: 20499}},
	"south_carolina": {Host: "state.decodo.com", RotatingPort: 20500, StickyPortRange: PortRange{Start: 20501, End: 20599}},
	"south_dakota":   {Host: "state.decodo.com", RotatingPort: 20600, StickyPortRange: PortRange{Start: 20601, End: 20699}},
	"tennessee":      {Host: "state.decodo.com", RotatingPort: 20700, StickyPortRange: PortRange{Start: 20701, End: 20799}},
	"texas":          {Host: "state.decodo.com", RotatingPort: 14000, StickyPortRange: PortRange{Start: 14001, End: 14999}},
	"utah":           {Host: "state.decodo.com", RotatingPort: 20800, StickyPortRange: PortRange{Start: 20801, End: 20899}},
	"vermont":        {Host: "state.decodo.com", RotatingPort: 20900, StickyPortRange: PortRange{Start: 20901, End: 20999}},
	"virginia":       {Host: "state.decodo.com", RotatingPort: 15000, StickyPortRange: PortRange{Start: 15001, End: 15999}},
	"washington":     {Host: "state.decodo.com", RotatingPort: 16000, StickyPortRange: PortRange{Start: 16001, End: 16999}},
	"west_virginia":  {Host: "state.decodo.com", RotatingPort: 21000, StickyPortRange: PortRange{Start: 21001, End: 21099}},
	"wisconsin":      {Host: "state.decodo.com", RotatingPort: 21100, StickyPortRange: PortRange{Start: 21101, End: 21199}},
	"wyoming":        {Host: "state.decodo.com", RotatingPort: 21200, StickyPortRange: PortRange{Start: 21201, End: 21299}},
}

// Preset returns the endpoint preset for the configured targeting, or a false ok return if none match.
func (c Config) Preset() (EndpointPreset, bool) {
	if c.EndpointSpec.Host != "" {
		return EndpointPreset{}, false
	}

	city := normalizeToken(c.Targeting.City)
	state := normalizeToken(c.Targeting.State)
	country := normalizeToken(c.Targeting.Country)

	// City takes precedence over state
	if city != "" {
		if preset, ok := cityPresets[city]; ok {
			return preset, true
		}
	}

	// State takes precedence over country for US states
	// Strip "us_" prefix if present since validation requires it but preset keys don't have it
	if state != "" {
		stateKey := strings.TrimPrefix(state, "us_")
		if preset, ok := statePresets[stateKey]; ok {
			return preset, true
		}
	}

	if country != "" {
		if preset, ok := wellKnownPresets[country]; ok {
			return preset, true
		}
	}

	return EndpointPreset{}, false
}

// ApplyPreset updates the EndpointSpec to match the targeting configuration
// using well-known Decodo endpoint presets. If no matching preset is found,
// no changes are made. This allows targeting to automatically select the
// correct endpoint, port, and sticky port range.
func (c *Config) ApplyPreset() {
	if preset, ok := c.Preset(); ok {
		c.EndpointSpec = EndpointSpec{
			Host:            preset.Host,
			RotatingPort:    preset.RotatingPort,
			StickyPortRange: preset.StickyPortRange,
		}
	}
}

type SessionType string

const (
	// SessionTypeRotating requests a new residential IP on each proxy request.
	SessionTypeRotating SessionType = "rotating"
	// SessionTypeSticky keeps the same residential IP for the configured session duration.
	SessionTypeSticky SessionType = "sticky"
)

// Config describes a Decodo user:pass backconnect proxy configuration.
type Config struct {
	Auth         Auth
	EndpointSpec EndpointSpec
	Endpoint     string
	Port         int
	Targeting    Targeting
	Session      Session
}

// Auth stores the raw Decodo proxy username and password from the dashboard.
type Auth struct {
	Username string
	Password string
}

// Targeting describes optional Decodo location and carrier targeting parameters.
type Targeting struct {
	Country   string
	City      string
	State     string
	ZIP       string
	Continent string
	ASN       int
}

// Session describes whether requests should rotate IPs or reuse a sticky session.
type Session struct {
	Type            SessionType
	ID              string
	DurationMinutes int
}

// EndpointSpec describes a Decodo endpoint together with its rotating port and sticky port range.
type EndpointSpec struct {
	Host            string
	RotatingPort    int
	StickyPortRange PortRange
}

// PortRange describes an inclusive port range.
type PortRange struct {
	Start int
	End   int
}

// TTL returns the sticky-session lifetime as a time.Duration.
func (s Session) TTL() time.Duration {
	if s.Type != SessionTypeSticky || s.DurationMinutes <= 0 {
		return 0
	}

	return time.Duration(s.DurationMinutes) * time.Minute
}

// NewEndpointSpec validates and returns a Decodo endpoint specification.
func NewEndpointSpec(host string, rotatingPort int, stickyPortRange PortRange) (EndpointSpec, error) {
	spec := EndpointSpec{
		Host:            strings.TrimSpace(strings.ToLower(host)),
		RotatingPort:    rotatingPort,
		StickyPortRange: stickyPortRange,
	}

	if err := spec.Validate(); err != nil {
		return EndpointSpec{}, err
	}

	return spec, nil
}

// Validate checks whether the endpoint specification is structurally valid.
func (e EndpointSpec) Validate() error {
	if e.IsZero() {
		return nil
	}

	if strings.TrimSpace(e.Host) == "" {
		return errors.New("endpoint spec host is required")
	}

	if e.RotatingPort <= 0 {
		return errors.New("endpoint spec rotating port must be positive")
	}

	return e.StickyPortRange.Validate()
}

// IsZero reports whether the endpoint specification is unset.
func (e EndpointSpec) IsZero() bool {
	return strings.TrimSpace(e.Host) == "" && e.RotatingPort == 0 && e.StickyPortRange.IsZero()
}

// Validate checks whether the port range is structurally valid.
func (r PortRange) Validate() error {
	if r.IsZero() {
		return nil
	}

	if r.Start <= 0 || r.End <= 0 {
		return errors.New("port range values must be positive")
	}

	if r.End < r.Start {
		return errors.New("port range end must be greater than or equal to start")
	}

	return nil
}

// IsZero reports whether the port range is unset.
func (r PortRange) IsZero() bool {
	return r.Start == 0 && r.End == 0
}

// Contains reports whether the range includes the provided port.
func (r PortRange) Contains(port int) bool {
	if r.IsZero() {
		return false
	}

	return port >= r.Start && port <= r.End
}

func (r PortRange) size() int {
	if r.IsZero() {
		return 0
	}

	return r.End - r.Start + 1
}

// NewAuth validates and normalizes raw Decodo dashboard credentials.
func NewAuth(username, password string) (Auth, error) {
	auth := Auth{
		Username: strings.TrimSpace(username),
		Password: strings.TrimSpace(password),
	}

	if err := auth.Validate(); err != nil {
		return Auth{}, err
	}

	return auth, nil
}

// Validate checks whether the credentials can be used to build a Decodo proxy username.
func (a Auth) Validate() error {
	if strings.TrimSpace(a.Username) == "" {
		return errors.New("username is required")
	}

	if strings.HasPrefix(strings.TrimSpace(a.Username), "user-") {
		return errors.New("username must be the raw decodo proxy username without the user- prefix")
	}

	if strings.TrimSpace(a.Password) == "" {
		return errors.New("password is required")
	}

	return nil
}

// Validate checks whether the configuration satisfies Decodo parameter constraints.
func (c Config) Validate() error {
	normalized, err := c.Normalized()
	if err != nil {
		return err
	}

	if err := normalized.Auth.Validate(); err != nil {
		return err
	}

	if err := normalized.EndpointSpec.Validate(); err != nil {
		return err
	}

	if normalized.Targeting.ASN > 0 {
		if normalized.Targeting.Country != "" || normalized.Targeting.City != "" || normalized.Targeting.State != "" || normalized.Targeting.ZIP != "" || normalized.Targeting.Continent != "" {
			return errors.New("asn cannot be combined with other targeting parameters")
		}
	}

	if normalized.Targeting.Continent != "" && (normalized.Targeting.Country != "" || normalized.Targeting.City != "" || normalized.Targeting.State != "" || normalized.Targeting.ZIP != "") {
		return errors.New("continent targeting cannot be combined with country, state, city, or zip")
	}

	if normalized.Targeting.City != "" && normalized.Targeting.Country == "" {
		return errors.New("city targeting requires country")
	}

	if normalized.Targeting.State != "" {
		if normalized.Targeting.Country != "us" {
			return errors.New("state targeting requires country us")
		}
		if !strings.HasPrefix(normalized.Targeting.State, "us_") {
			return errors.New("state targeting must use us_ prefix")
		}
	}

	if normalized.Targeting.ZIP != "" {
		if normalized.Targeting.Country != "us" {
			return errors.New("zip targeting requires country us")
		}
		if len(normalized.Targeting.ZIP) != 5 || !allDigits(normalized.Targeting.ZIP) {
			return errors.New("zip targeting requires a 5-digit zip")
		}
		if normalized.Targeting.City != "" || normalized.Targeting.State != "" || normalized.Targeting.Continent != "" {
			return errors.New("zip targeting cannot be combined with city, state, or continent")
		}
	}

	switch normalized.Session.Type {
	case "", SessionTypeRotating:
		if normalized.Session.DurationMinutes != 0 {
			return errors.New("rotating session cannot set duration")
		}
		if normalized.Session.ID != "" {
			return errors.New("rotating session cannot set session id")
		}
	case SessionTypeSticky:
		if normalized.Session.ID == "" {
			return errors.New("sticky session requires session id")
		}
		if normalized.Session.DurationMinutes < 1 || normalized.Session.DurationMinutes > 1440 {
			return errors.New("sticky session duration must be between 1 and 1440 minutes")
		}
	default:
		return fmt.Errorf("unsupported session type %q", normalized.Session.Type)
	}

	if normalized.Session.Type == SessionTypeSticky && !normalized.EndpointSpec.IsZero() && !normalized.EndpointSpec.StickyPortRange.IsZero() && !normalized.EndpointSpec.StickyPortRange.Contains(normalized.Port) {
		return errors.New("sticky session port must be inside the endpoint sticky port range")
	}

	return nil
}

// Normalized returns a copy of the configuration with defaults and normalized tokens applied.
func (c Config) Normalized() (Config, error) {
	normalized := c

	normalized.Auth.Username = strings.TrimSpace(normalized.Auth.Username)
	normalized.Auth.Password = strings.TrimSpace(normalized.Auth.Password)
	normalized.Endpoint = strings.TrimSpace(strings.ToLower(normalized.Endpoint))
	normalized.EndpointSpec.Host = strings.TrimSpace(strings.ToLower(normalized.EndpointSpec.Host))

	normalized.Targeting.Country = normalizeToken(normalized.Targeting.Country)
	normalized.Targeting.City = normalizeToken(normalized.Targeting.City)
	normalized.Targeting.State = normalizeToken(normalized.Targeting.State)
	normalized.Targeting.ZIP = strings.TrimSpace(normalized.Targeting.ZIP)
	normalized.Targeting.Continent = normalizeToken(normalized.Targeting.Continent)

	normalized.Session.ID = strings.TrimSpace(normalized.Session.ID)
	if normalized.Session.Type == "" {
		if normalized.Session.ID != "" {
			normalized.Session.Type = SessionTypeSticky
		} else {
			normalized.Session.Type = SessionTypeRotating
		}
	}

	if normalized.Session.Type == SessionTypeSticky && normalized.Session.DurationMinutes == 0 {
		normalized.Session.DurationMinutes = defaultStickyDurationMinutes
	}

	if normalized.EndpointSpec.IsZero() {
		if normalized.Endpoint == "" {
			normalized.Endpoint = defaultEndpoint
		}
		if normalized.Port == 0 {
			normalized.Port = defaultPort
		}
	} else {
		if normalized.Endpoint == "" {
			normalized.Endpoint = normalized.EndpointSpec.Host
		}
		if normalized.Port == 0 {
			if normalized.Session.Type == SessionTypeSticky && !normalized.EndpointSpec.StickyPortRange.IsZero() {
				normalized.Port = normalized.EndpointSpec.StickyPortRange.Start
			} else {
				normalized.Port = normalized.EndpointSpec.RotatingPort
			}
		}
	}

	if err := normalized.ValidateShallow(); err != nil {
		return Config{}, err
	}

	return normalized, nil
}

// ProxyUsername builds the Decodo proxy username, including targeting and session parameters.
func (c Config) ProxyUsername() (string, error) {
	normalized, err := c.Normalized()
	if err != nil {
		return "", err
	}

	if err := normalized.Validate(); err != nil {
		return "", err
	}

	parts := []string{"user", normalized.Auth.Username}

	if normalized.Targeting.Continent != "" {
		parts = append(parts, "continent", normalized.Targeting.Continent)
	}
	if normalized.Targeting.Country != "" {
		parts = append(parts, "country", normalized.Targeting.Country)
	}
	if normalized.Targeting.State != "" {
		parts = append(parts, "state", normalized.Targeting.State)
	}
	if normalized.Targeting.City != "" {
		parts = append(parts, "city", normalized.Targeting.City)
	}
	if normalized.Targeting.ZIP != "" {
		parts = append(parts, "zip", normalized.Targeting.ZIP)
	}
	if normalized.Targeting.ASN > 0 {
		parts = append(parts, "asn", strconv.Itoa(normalized.Targeting.ASN))
	}
	if normalized.Session.Type == SessionTypeSticky {
		parts = append(parts, "session", normalized.Session.ID, "sessionduration", strconv.Itoa(normalized.Session.DurationMinutes))
	}

	return strings.Join(parts, "-"), nil
}

// ProxyURL builds an authenticated Decodo proxy URL suitable for HTTP proxy clients.
func (c Config) ProxyURL() (*url.URL, error) {
	normalized, err := c.Normalized()
	if err != nil {
		return nil, err
	}

	username, err := normalized.ProxyUsername()
	if err != nil {
		return nil, err
	}

	return &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", normalized.Endpoint, normalized.Port),
		User:   url.UserPassword(username, normalized.Auth.Password),
	}, nil
}

// ValidateShallow checks lightweight structural constraints before full validation.
func (c Config) ValidateShallow() error {
	if c.Port < 0 {
		return errors.New("port must be positive")
	}
	return nil
}

func normalizeToken(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, " ", "_")
	return value
}

func allDigits(value string) bool {
	for _, r := range value {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
