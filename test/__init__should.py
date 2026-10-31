# __init__.py was generated from init.ct.
# if you edit __init__.py, run `ct -tc __init__.py` afterwards to update init.ct.
# when editing this file leave the `` tags.
#``ct version 0.1.1 
#``ct files init.ct __init__.py 

# # traction
# 
# commonly used getters (and setters?) for centraxx db.  1``//__init__.py: #py
# import dbcq and cnf.  10``/import: #py
from dbcq import dbcq
import cnf
# import.  49``/import
from dip import dig
from tram import Sample, Idable, Amount, Identifier
# import.  77``/import
from tram import Patient
# import.  111``/import
from tram import Finding
from tram import Rec, BooleanRec, NumberRec, StringRec, DateRec, MultiRec, CatalogRec
# import.  158``/import
from tram import User
from tram import Address
# import re.  312``/import
import re
# import csv and sys.  348``/import
import csv
import sys
#``10
# at the package level are a series of constants that can be used in
# function calls, e.g. sample(..., verbose=[tr.parentid,
# tr.locationpath]).
# 
# maybe keep the names in lowercase so that they are the same as the
# keys in the returned json?  2``/const: #py
address = "address"
appointment = "appointment"
catalog = "catalog"
category = "category" # MASTER, ALIQUOTGROUP, DERIVED. dtype in db.
concentration = "concentration"
cxxkitid = "cxxkitid"
creationdate = "creationdate" 
derivaldate = "derivaldate" # aufteilungsdatum / date of distribution
email = "email"
initialamount = "initialamount"
initialunit = "initialunit"
first_repositiondate = "first_repositiondate" # datum der ersten einlagerung / date of first storage (not in fhir). is identical to derivaldate. first_repositiondate in db.
method = "method"
kitid = "kitid"
labval = "labval"
lastlogin = "lastlogin"
login = "login"
locationname = "locationname"
locationpath = "locationpath"
orga = "orga"
parentid = "parentid"
parentoid = "parentoid"
patientid = "patientid"
project = "project"
receiptdate = "receiptdate" # eingangsdatum / date of receipt. receiptdate in db.
receptacle = "receptacle"
repositiondate = "repositiondate" # datum der letzten einlagerung / most recent storage date. 
restamount = "restamount"
restunit = "restunit"
sampleid = "sampleid"
sampleoid = "sampleoid"
samplingdate = "samplingdate" # entnahmedatum / extraction date.
secondprocessing = "secondprocessing"
secondprocessingdate = "secondprocessingdate"
stockprocessing = "stockprocessing"
stockprocessingdate = "stockprocessingdate"
trial = "trial"
type = "type" # sampletype (EDTA, stool etc) in db.
username = "username"
values = "values"
xposition = "xposition"
yposition = "yposition"
# add a variable that holds the conf template.  11``/const
cnftemplate = """
# settings for traction.

# sampleid sets the idcontainertype code that is used when searching for sampleid.
# put in a code per db target.
sampleid: 
  <db target>: <an idcontainertype code, e.g. SAMPLEID>
# patientid sets the idcontainertype code that is used when searching for patientid.
# put in a code per db target.
patientid: 
  <db target>: <an idcontainertype code, e.g. LIMSPSN>

# idc holds additional idcontainertype codes that will be queryable as command line flags.
idc:
 - <an idcontainertype code>
 - <another idcontainertype code>

# cxx holds the centraxx version for db target
cxx:
  <db target>: 3|4
"""
#``2

# _checkverbose makes shure that only allowed keys are in the verbose array.   287``/_checkverbose:
def _checkverbose(verbose, possible):
    for verb in verbose:
        if verb not in possible: 
            print(f"error: verbose entry {verb} must be in {possible}.")
            return False
    return True
#``287
# floatornull casts to float or returns None. somehow the values passed
# to Amount need float casts, and float casts don't accept None.  308``/floatornull:
def floatornull(x):
    if x is None:
        return None
    return float(x)
#``308
# get_ids returns a list of string ids from a list of Idables (Sample or
# Patient).  309``/get_ids:
def get_ids(idables:list, code:str|None=None) -> list:
    return [ x.id(code) for x in idables ]
#``309
# isnumber returns if a string is a number to prevent sql injection.  is
# this handled better by an existing library function?  310``/isnumber:
def isnumber(a) -> bool:
    return re.match(r"^[0-9](.[0-9]*)?$", a) is not None
#``310
# isidentifier returns whether a string is a sql identifier to prevent
# sql injection.  like isnumber, could this be handled better by a library?  311``/isidentifier:
def isidentifier(a) -> bool:
    return re.match(r"^[A-Za-z_0-9\.#]+$", a) is not None
#``311
# _dextends extends an array in a dictionary at the given key with the
# values in the passed list.  313``/_dextend:
def _dextend(d:dict, key, lst:list|None):
    # the list should hold values.  314.``
    if lst is not None:
        if key not in d:
            d[key] = []
        d[key].extend(lst)
    #``314
#``313
# _tablelists returns a dict of the lists smaller than cutoff and a
# second dict of the lists larger than cutoff that go into temporary
# tables.  315``/_tablelists:
def _tablelists(all:dict, cutoff:int): # -> (dict, dict)
    # the two out dicts.  316.``
    outsmall = {}
    outlarge = {}
    # go over dicta, if a list is larger than cutoff, move it to dictb and
    # delete it from dicta.  317``
    for key, lst in all.items():
        if lst is not None and len(lst) > cutoff:
            outlarge[key] = lst
        else:
            outsmall[key] = lst
    # return.  318``
    return outsmall, outlarge
    #``316
#``315
# _uniq returns a new list containing the unique items of the list passed, preserving order.  319``/_uniq:
def _uniq(lst): # -> list
    out = []
    for e in lst:
        if e not in out:
            out.append(e)
    return out
#``319
# _byifone sets by to the only passed query parameter, if missing is
# true and there is only one query parameter.  320``/_byifone:
def _byifone(alllists):
    throwtogether = list(alllists["idc"].keys()) + list(alllists["nonidc"].keys())
    if len(throwtogether) == 1:
        return throwtogether[0]
    else:
        raise Exception("--missing needs a tr constant or idc passed to --by")
#``320
# _prepby returns a bydict that holds the query params for the by as key
# (if there are any), and a second list, missinglst, that holds these
# query params so that they can subsequently be deleted so that the
# truly missing remain.  321``/_prepby:
def _prepby(alllists:dict, by:str=None, missing:bool=None):
    # make a bydict and missinglst.  322``_:
    bydict = None
    missinglst = None
    if by is not None:
        # fill the missing list, if missings should be returned.  323``./missing:
        if missing:
            missinglst = []
            # todo wrap this in function: missinglst =  idc_or_not(alllists, by)
            if by in alllists["idc"]:
                missinglst = alllists["idc"][by]
            elif by in alllists["nonidc"]:
                missinglst = alllists["nonidc"][by]
        #``323
        # if query params were passed for this by, they become keys in the
        # returned by-dict, this way query params for which there are no samples
        # can be found.  324``../bydict:
        else:
            bydict = {}
            if by in alllists["idc"]:
                for byval in alllists["idc"][by]:
                    bydict[byval] = []
            if by in alllists["nonidc"]:
                for byval in alllists["nonidc"][by]:
                    bydict[byval] = []
        #``324
    # return.  325``/_prepby/_
    return bydict, missinglst
    #``322
#``321
# _fillby puts a value into the bydict at the key of the by-field.  326``/_fillby:
def _fillby(bydict, by, row, what):
    # for now, idcs are just fields in the result row (dbcq makes them
    # lowercase), so fetching them from the row should work for both tr
    # constants and idcs.  327.``
    if bydict is not None:
        bykey = dig(row, by.lower())
        if bykey is not None and not bykey in bydict:
            bydict[bykey] = []
        bydict[bykey].append(what)
    #``327
#``326
# _updatemissing removes the value at the by-field from the missing list.  328``/_updatemissing:
def _updatemissing(missinglst, by, row):
    # for now, idcs are just fields in the result row (dbcq makes them
    # lowercase), so fetching them from the row should work for both tr
    # constants and idcs.  329.``
    bykey = dig(row, by.lower())
    if bykey in missinglst:
        missinglst.remove(bykey)
    #``329
#``328

# dict_csv writes a list of dicts to csv file. if no column names are
# given, the fields in the first dict are used as column names.  330``/dict_csv:
def dict_csv(lst:list, colnames=None, outfile=None, delim:str=",") -> str|None:
    # arg parse passes None for delim if it wasn't set, set it here.  331.``
    if delim is None:
        delim = ","
    # if no list return.   332``
    if lst is None or len(lst) == 0: # todo throw error?
        return None
    # if a path is passed, open the file, else assume it's sys.stdout.  333``
    out = sys.stdout 
    if isinstance(outfile, str):
        out = open(outfile, "w", newline="")
    # make a csv dict writer.
    # 
    # take the fields of the first list entry as column names.  334``
    if colnames is None:
        colnames = list(lst[0].keys())
    
    writer = csv.DictWriter(out, fieldnames=colnames, delimiter=delim) 
    writer.writeheader()
    # write a row for each dict.  335``
    for d in lst:
        writer.writerow(d)
    # close and return the out file.  336``
    out.close()
    return outfile
    #``331
#``330
# flat_csv writes a list of 'flat' tram objects like User (that don't
# need to pull in nested data into their dict root) to csv file.  337``/flat_csv:
def flat_csv(lst:list, outfile=None, delim:str=",") -> str|None:
    # arg parse passes None for delim if it wasn't set, set it here.  338.``
    if delim is None:
        delim = ","
    # if no list return.   339``
    if lst is None or len(lst) == 0: # todo throw error?
        #print("no idables")
        return None
    # put the dict of each element into an array for dict_csv.  340``
    rows = []
    for e in lst:
        rows.append(e.__dict__)
    # send the rows to dict_csv and return the written filename, if any.  341``
    return dict_csv(rows, outfile=outfile, delim=delim)    
    #``338
#``337
# idable_csv writes a list of Idables to the given csv file. the given
# idcontainers become columns in the csv. if no idcontainers are given,
# all are included. if True is passed as file, the output is
# printed. (todo pass sys.stdout instead?)  342``/idable_csv:
def idable_csv(idables:list, outfile=None, delim:str=",", *idcs) -> str|None:
    # arg parse passes None for delim if it wasn't set, set it here.  343.``
    if delim is None:
        delim = ","
    # if no idables return.   344``
    if idables is None or len(idables) == 0: # todo throw error?
        #print("no idables")
        return None
    # collect the column names, since we don't know from the get-go which
    # idcontainers will be there.  345``
    colnames = []
    for idable in idables:
        for col in list(idable.iddict(*idcs).keys()):
            if col not in colnames:
                colnames.append(col)
    # get the dict for each idable.  346``
    rows = []
    for idable in idables:
        d = idable.iddict(*idcs)
        rows.append(d)
    # send the rows to dict_csv and return the written filename, if any.  347``
    return dict_csv(rows, colnames=colnames, outfile=outfile, delim=delim)    
    #``343
#``342
# finding_csv writes a list of Findings along their recorded values to
# the given csv file, the recorded values in the same row as their
# respective finding. if True is passed as file the output is printed. todo use sys.stdout instead of True.  349``/finding_csv:
def finding_csv(findings:list, outfile=None, delim:str=",", delim_cmp:str=",") -> str:
    # arg parse passes None for delim if it wasn't set, set it here.  350.``
    if delim is None:
        delim = ","
    if delim_cmp is None:
        delim_cmp = ","
    # if no findings return.   351``
    if findings is None or len(findings) == 0: # todo throw error?
        print("no findings")
        return None
    # make dicts from the findings and pull in the recorded values.
    # 
    # collect a list of all the csv column names, so that values end up in
    # the right columns when writing csv.  352``
    fdicts = []
    colnames = []

    for finding in findings:
        # get the dict for the finding, delete the recs field.  353``./build fdict:
        fdict = finding.__dict__.copy()
        del fdict["recs"]
        # if it's the first row, remember the finding field names for the csv
        # header. they should be the same for all findings.  354``
        if len(colnames) == 0:
            colnames.extend(list(fdict.keys()))
        # pull in the recorded values.  355``
        for code, rec in finding.recs.items():
            # for each recorded value, make a field cmp_t_CODE holding the type and
            # cmp_v_CODE holding the value(s).
            # 
            # findings come with different recorded values, if this recorded value's
            # labval hasn't been encountered yet, remember it as column name.  356``./pull rec:
            tkey = f"cmp_t_{rec.labval}"
            vkey = f"cmp_v_{rec.labval}"
            if tkey not in colnames:
                colnames.append(tkey)
            if vkey not in colnames:
                colnames.append(vkey)
            # go through the recorded values by type.  357``
            if isinstance(rec, BooleanRec):
                fdict[tkey] = "BOOL"
                fdict[vkey] = rec.rec
            # for number.  358``
            if isinstance(rec, NumberRec):
                fdict[tkey] = "NUMBER"
                fdict[vkey] = rec.rec
            # for string.  359``
            if isinstance(rec, StringRec):
                fdict[tkey] = "STRING"
                fdict[vkey] = rec.rec
            # for date.  360``
            if isinstance(rec, DateRec):
                fdict[tkey] = "DATE"
                fdict[vkey] = rec.rec
            # for multi and catalog rec, put all entries in its rec list into one field.  361``
            if isinstance(rec, MultiRec):
                fdict[tkey] = "MULTI"
                fdict[vkey] = delim_cmp.join(rec.rec)
            if isinstance(rec, CatalogRec):
                fdict[tkey] = "CATALOG"
                fdict[vkey] = delim_cmp.join(rec.rec)
                
            #``356
        # after adding the recorded values append the finding dict.  362``../
        fdicts.append(fdict)
        # send the ficts to dict_csv and return the written filename, if any.  363``
    return dict_csv(fdicts, colnames=colnames, outfile=outfile, delim=delim)    
        #``353
    #``350
#``349
# method_csv writes a list of Methods to the given csv file, one row per labval in the method. if True is passed as file the output is printed. todo use sys.stdout instead of True.  364``/method_csv:
def method_csv(methods:list, outfile=None, delim:str=",", delim_usageentry:str=",") -> str:
    # arg parse passes None for delim if it wasn't set, set it here.  365``_:
    if delim is None:
        delim = ","
    if delim_usageentry is None:
        delim_usageentry = ","
    # if no methods return.   366``
    if methods is None or len(methods) == 0: # todo throw error?
        return None
    # make dicts for each labval.
    # 
    # collect a list of all the csv column names, so that values end up in
    # the right columns when writing csv.  367``
    rows = []
    colnames = { "method": None, "method_de": None, "method_en": None, "labval": None, "labval_en": None, "labval_de": None, "labval_type": None }

    for method in methods:
        for labval in method["labvals"].values():
            # make a row for each labval.  368.``
            row = {}
            # put in the method code.  369``
            row["method"] = method["code"]
            row["method_de"] = method["name_de"]
            row["method_en"] = method["name_en"]
            # put in the labval code and type.  370``
            row["labval"] = labval["code"]
            row["labval_de"] = labval["name_de"]
            row["labval_en"] = labval["name_en"]
            row["labval_type"] = labval["type"]
            # put in catalog code and associated usageentries.  371``
            if "catalog" in labval:
                if "labval_catalog" not in colnames:
                    colnames["labval_catalog"] = None
                row["labval_catalog"] = labval["catalog"]
            if "usageentry" in labval:
                if "labval_usageentry" not in colnames:
                    colnames["labval_usageentry"] = None
                row["labval_usageentry"] = delim_usageentry.join(list(labval["usageentry"].keys()))
            # add.  372``
            rows.append(row)
            #``368
    # write the csv.  373``/method_csv/_
    return dict_csv(rows, colnames=list(colnames.keys()), outfile=outfile, delim=delim)
    #``365
#``364
# catalog_csv writes a list of catalog dicts to the given csv file, one
# row per entry in the catalog. if True is passed as file the output is
# printed. todo use sys.stdout instead of True.  374``/catalog_csv:
def catalog_csv(catalogs:list, outfile=None, delim:str=",") -> str:
    # arg parse passes None for delim if it wasn't set, set it here.  375``_:
    if delim is None:
        delim = ","
    # if no catalogs return.   376``
    if catalogs is None or len(catalogs) == 0: # todo throw error?
        return None
    # make the rows.  377``
    rows = []
    for catalog in catalogs:
        for entry in catalog["entries"].values():
            # make a row for each entry.  378.``
            row = {}
            # put in the catalog info.   379``
            row["catalog"] = catalog["code"]
            row["catalog_de"] = catalog["name_de"]
            row["catalog_en"] = catalog["name_en"]
            # put in the entry info.  380``
            row["entry"] = entry["code"]
            row["entry_de"] = entry["name_de"]
            row["entry_en"] = entry["name_en"]
            # add.  381``
            rows.append(row)
            #``378
    # write the csv. column names need not be passed, cause they are the same for each row.  382``/catalog_csv/_
    return dict_csv(rows, outfile=outfile, delim=delim)
    #``375
#``374

class traction:
    # __init__ takes the db target either as string or dbcq object.  3``/init: #py
    def __init__(self, target):
        # get the settings. if the settings file is missing, cnf creates it and
        # asks the user to edit it.  4.``
        self.settings = cnf.makeload(path=".traction/settings.yaml", root=cnf.home, fmt="yaml", make=cnftemplate)        
        # if no settings, do nothing.   5``
        if self.settings is None:
            raise Exception("no settings.")
            return
        # either pass a string argument to dbcq or use dbcq instance directly.  6`` #py
        if isinstance(target, str):
            self.db = dbcq(target)
        elif isinstance(target, dbcq): 
            self.db = target
        else:
            raise Exception("target needs to be string or dbcq instance")
        # cache the id containers' oids and kinds by code.  7``
        self._idcinit()
        # print(f"_idcoids: {self._idcoids}")
        # make a join-dict, jd, that holds joins from table a to table b.
        # 
        # init jd here, cause it needs to reference self at one point for joining in an idc_ table.  8``
        self.jd = {
            # the joindict entries that sample needs.  50``/init/jd sample_to:
            "sample_to_samplekit": [f"left join centraxx_samplekititem as samplekititem on samplekititem.tubebarcode = idc_{self.sidc()}.psn", "left join centraxx_samplekit as samplekit on samplekit.oid = samplekititem.samplekit"],
            "sample_to_parentid": [f"left join centraxx_sampleidcontainer parentidc on parentidc.sample = sample.parent and parentidc.idcontainertype={self._idcoid[self.sidc()]}"],
            "sample_to_samplelocation": ["left join centraxx_samplelocation samplelocation on samplelocation.oid = sample.samplelocation"],
            "sample_to_sampletype": ["left join centraxx_sampletype as sampletype on sampletype.oid = sample.sampletype"],
            "sample_to_stockprocessing": ["left join centraxx_stockprocessing as stockprocessing on sample.stockprocessing = stockprocessing.oid"],
            "sample_to_secondprocessing": ["left join centraxx_stockprocessing as secondprocessing on sample.secondprocessing = secondprocessing.oid"],
            "sample_to_project": ["left join centraxx_project as project on sample.project = project.oid"],
            "sample_to_patient": ["left join centraxx_patientcontainer as patientcontainer on sample.patientcontainer = patientcontainer.oid"],
            "sample_to_receptacle": ["left join centraxx_samplereceptable as receptable on sample.receptable = receptable.oid"], # receptable seems to be a typo in the table naming
            "sample_to_orga": ["left join centraxx_organisationunit as organisationunit on sample.orgunit = organisationunit.oid"],
            "sample_to_trial": ["left join centraxx_flexistudy as flexistudy on sample.flexistudy = flexistudy.oid"]
            #``50
            ,
            # the joindict entries that patient needs.  78``/init/jd patient_to:
            "patient_to_orga": ["left join centraxx_patientorgunit patientorgunit on patientcontainer.oid=patientorgunit.patientcontainer_oid", "left join centraxx_organisationunit organisationunit on patientorgunit.orgunit_oid=organisationunit.oid"],
            "patient_to_trial": ["left join centraxx_patientstudy as patientstudy on patientstudy.patientcontainer = patientcontainer.oid", "left join centraxx_flexistudy as flexistudy on flexistudy.oid = patientstudy.flexistudy"],
            "patient_to_sample": ["left join centraxx_sample sample on sample.patientcontainer = patientcontainer.oid"]
            #``78
            ,
            # the join dict entries for user.  157``/init/jd participant_to:
            "participant_to_address": ["left join centraxx_participantaddress participantaddress on participantaddress.participant = participant.oid", "left join centraxx_address address on address.oid = participantaddress.oid"],
            "participant_to_credential": ["left join centraxx_credential credential on credential.participant = participant.oid"]
            #``157
            ,
            # the joindict entries that usageentry needs.  190``/init/jd usageentry_to:
            "usageentry_to_labval": [ "left join centraxx_labvalenum_usageentry labvalenum_usageentry on labvalenum_usageentry.usageentry = usageentry.oid", "left join centraxx_laborvalue labval on labvalenum_usageentry.labvalueenum = labval.oid" ]
            #``190
        }
        # set the display name caches.  9``
        self.names_labval = None
        self.names_catalogentry = None
        self.names_usageentry = None
        #``4
    #``3

    # sample gets sample(s) and returns them as a list of Sample instances.
    # 
    # pass sampleids and other values to filter for as lists of strings,
    # e.g. `sampleids=["a", "b", "c"]`.
    # 
    # pass idcontainer lists to idc, like `idc={"extsampleid": ["a", "b", "c"]}`.
    # 
    # pass fields that should be joined in into the verbose array, e.g.
    # `verbose=[tr.locationpath]`. to join in everything, say
    # `verbose_all=True`. this is slower than non-verbose.
    # 
    # pass dates as a tuple of from and to datetime, e.g. `samplingdates=(datefrom, None)`.
    # 
    # to match via like as opposed to exact, pass the respective tr
    # constants or idcs to like, e.g. `like=[tr.locationpath]`.
    # 
    # pass `by` to key returned samples by that tr constant or idc. if
    # missing is also passed, only missings are returned.  12``/sample: #py
    def sample(self, sampleids:list|None=None, oids:list|None=None, idc:dict|None=None, patientids:list|None=None, pidc:str|None=None, parentids:list|None=None, parentoids:list|None=None, locationpaths:list|None=None, locationnames:list|None=None, trials:list|None=None, kitids:list|None=None, cxxkitids:list|None=None, categories:list|None=None, types:list|None=None, orgas:list|None=None, samplingdates:list|None=None, receiptdates:list|None=None, derivaldates:list|None=None, first_repositiondates:list|None=None, repositiondates:list|None=None, stockprocessingdates:list|None=None, secondprocessingdates:list|None=None, files:dict|None=None, verbose:list|None=None, verbose_all:bool=False, primaryref:bool=False, incl_parents:bool=False, incl_childs:bool=False, incl_tree:bool=False, like:list|None=None, missing=False, order_by:str|None=None, top:int|None=None, by=None, print_query:bool=False, raw:bool=False):
        # init arrays.  don't set them as mutable default parameters
        # (e.g. verbose=[]), cause they would then retain their values across
        # method calls. see
        # https://www.geeksforgeeks.org/python/use-mutable-default-value-as-an-argument-in-python/  13``_:
        if verbose is None:
            verbose = []
        if like is None:
            like = []
        if files is None:
            files = {}
        if idc is None:
            idc = {}
        # ## took code from here.
        # 
        # walk the sample tree if needed.
        # 
        # there are two cases where we walk the sample tree:
        # 
        # (a) including other samples.
        # 
        # if the sample's parents, children or it's whole tree should be
        # included in the result, first run this call, take the sample oids
        # returned, then add the oids of the parents / childs / tree samples (in
        # order from root to leaf), and then run the query with its verboses for
        # all the sample oids.
        # 
        # (b) referencing the primary sample for each derived sample.
        # 
        # if each derived sample should reference its primary, also
        # run the call, and then fill in the primary field for each sample.
        # 
        # at first it was thought to do the primary referencing via verbose, but
        # then each verbose all call would include the primary references, which
        # might slow it down?
        # 
        # check if we walk the sample tree.  14``
        if incl_parents or incl_childs or incl_tree or primaryref:
            # we fire the call with all the arguments except incl_*, and raw.
            # 
            # we pass verbose only if we add primary references, cause then we only
            # run the call once. if we include tree samples later we don't pass
            # verbose, cause we eventually run the call again for all tree samples,
            # passing verbose here would slow it down.
            # 
            # it shouldn't make a difference to pass verbose_all for primary, cause
            # verbose should be filled in any case.  15``/sample/_/incl:
            verbosepass = []
            if primaryref:
                verbosepass = verbose
            # fire the call, then add to the result.  16``
            res = self.sample(sampleids=sampleids, idc=idc, parentids=parentids, parentoids=parentoids, patientids=patientids, pidc=pidc, trials=trials, locationpaths=locationpaths, kitids=kitids, cxxkitids=cxxkitids, categories=categories, samplingdates=samplingdates, receiptdates=receiptdates, derivaldates=derivaldates, first_repositiondates=first_repositiondates, repositiondates=repositiondates, stockprocessingdates=stockprocessingdates, secondprocessingdates=secondprocessingdates, verbose=verbosepass, verbose_all=False, like=like, missing=missing, order_by=order_by, top=top, print_query=print_query)  ## why is verbose_all here False? TODO should be verbose_all=verbose_all if we didn't fill verbose until now?
            # proceed differently if primary info should be added or if samples from
            # the tree should be included.  17``
            if primaryref:
                for sample in res:
                    self._fill_in_primary(sample)
                return res
            # when including samples from the tree, take all the sample oids.
            # 
            # use oids to support aliquotgroups.  18``/sample/_/incl
            
            s_oids = get_ids(res, "oid")
            # withincl holds all the sampleids that eventually get queried.  19``
            withincl = []
            # now put in the sample oids of parents, childs or the whole tree.  20``
            for s_oid in s_oids:
               if incl_parents:
                   # if the parent sampleids should be included, get them from root to
                   # leaf. collect them into the pids list.  21``/sample/_/incl/parents:
                   p_oids = []
                   self._get_parents(s_oid, p_oids)
                   # put the parent oids in, followed by our sample oid.  22``
                   withincl.extend(p_oids)
                   withincl.append(s_oid)
                   #``21
               if incl_childs:
                   # if the child oids should be included, get them from sample to leaf.  23``/sample/_/incl/childs:
                   c_oids = []
                   self._get_childs(s_oid, c_oids)
                   # put the child oids in, preceeded by our sample oid.  24``
                   withincl.append(s_oid)
                   withincl.extend(c_oids)
                   #``23
               if incl_tree:
                   # if the whole tree should be included, get the root, and include all
                   # children from there.  25``/sample/_/incl/tree:
                   p_oids = []
                   self._get_parents(s_oid, p_oids)
                   if len(p_oids) > 0:
                       root = p_oids[0]
                   else:
                       root = s_oid
                   # get the child oids from root.  26``
                   c_oids = []
                   self._get_childs(root, c_oids)
                   # append the root oid, then the child oids.  27``
                   withincl.append(root)
                   withincl.extend(c_oids)
                   #``25
            # after filling withincl, filter out recurring sampleids. keep each id
            # where it first appeared in the array to preserve tree dependencies.  28``/sample/_/incl
            withincl = list(dict.fromkeys(withincl))
            # using dict.fromkeys here depends on python dicts preserving insertion order
            # since python 3.7.
            # 
            # now run the query for all the sampleids with the verboses. the
            # verboses should contain all the query parameters from the original
            # call. in the call here we don't pass the query parameters from the
            # original call, cause we used them already to get the sampleids, and
            # here they could filter out parent/child/tree samples we'd like to keep.  29``
            return self.sample(oids=withincl, verbose=verbose, print_query=print_query, raw=raw) # todo pass verbose_all?
            #``15
        # these are all possible verbose options (verbose-all-array).
        # 
        # // let sampleid be the first element in vaa so it gets joined in first
        # for subsequent joins that depend on it.  30``/sample/_ #py
        vaa = [cxxkitid, kitid, locationname, locationpath, orga, parentid,
	       patientid, project, receptacle, type,
               secondprocessing, stockprocessing, trial]
        # append the idcs for patient and sample to the vaa array.    31``
        vaa.extend(self._pidcs(pidc))
        vaa.extend(self._sidcs())                
        # always join in the sampleid, as the first array element for subsequent
        # joins that might need it.  32`` #py
        if self.sidc() not in verbose:
            verbose.insert(0, self.sidc())
        # put the arguments in a dict of lists.  33``
        lists = {
          sampleoid: oids,
          parentid: parentids,
          parentoid: parentoids,
          locationpath: locationpaths,
          locationname: locationnames,
          trial: trials,
          kitid: kitids,
          cxxkitid: cxxkitids,
          category: categories,
          orga: orgas,
          samplingdate: samplingdates,
          receiptdate: receiptdates,
          derivaldate: derivaldates,
          first_repositiondate: first_repositiondates,
          repositiondate: repositiondates,
          stockprocessingdate: stockprocessingdates,
          secondprocessingdate: secondprocessingdates,
          type: types
        }
        # jselect holds statements for optional joins by key.
        # 
        # rename the field of the main sample idcontainer to sampleid.  34`` #py
        jselects = {
            self.sidc(): [f"idc_{self.sidc()}.psn as '{sampleid}'"],
            self.pidc(pidc): [f"idc_{self.pidc(pidc)}.psn as '{patientid}'"],                   
            cxxkitid: [f"samplekit.cxxkitid as '{cxxkitid}'"],
            #sampleid: [f"sidc.psn as '{sampleid}'"],
            parentid: [f"parentidc.psn as '{parentid}'"],
            kitid: [f"samplekit.kitid as '{kitid}'"],
            locationname: [f"samplelocation.locationid as '{locationname}'"], 
            locationpath: [f"samplelocation.locationpath as '{locationpath}'"],
            type: [f"sampletype.code as '{type}'"], # is there a type field already?
            stockprocessing: [f"stockprocessing.code as '{stockprocessing}'"],
            secondprocessing: [f"secondprocessing.code as '{secondprocessing}'"],
            project: [f"project.code as '{project}'"],
            #patientid: [f"patidc.psn as '{patientid}'"],
            receptacle: [f"receptable.code as '{receptacle}'"],
            orga: [f"organisationunit.code as '{orga}'"],
            trial: [f"flexistudy.code as '{trial}'"],
        }
        # lselects holds the local selects for this table, to rename fields.
        # 
        # put sample.* first, that it doesn't overwrite field of the renames?  35``
        lselects = [
          f"sample.amountrest as {restamount}",
          f"sample.appointmentnumber as {appointment}",
          f"sample.dtype as {category}",
          f"sample.oid as {sampleoid}",
          f"sample.parent as {parentoid}",
          "sample.*"
        ]
        # joins holds the join statements needed. if the same join is
        # needed for two different keys (e.g. locationname and locationpath),
        # filter them out later.  36`` #py
        joins = {
            cxxkitid: self.jd["sample_to_samplekit"],
            parentid: self.jd["sample_to_parentid"],
            kitid: self.jd["sample_to_samplekit"],
            locationname: self.jd["sample_to_samplelocation"],
            locationpath: self.jd["sample_to_samplelocation"],
            type: self.jd["sample_to_sampletype"],
            stockprocessing: self.jd["sample_to_stockprocessing"],
            secondprocessing: self.jd["sample_to_secondprocessing"],
            project: self.jd["sample_to_project"],
            receptacle: self.jd["sample_to_receptacle"],
            orga: self.jd["sample_to_orga"],
            trial: self.jd["sample_to_trial"]
        }
        # for any patient idc, join in patientcontainer, which is needed by
        # _join_idc.  37``
        for pidc in self._pidcs(pidc):
            joins[pidc] = self.jd["sample_to_patient"]
        # build the query and get the result.  38``
        (res, bydict, missinglst, by) = self._query(tablename="sample", lists=lists, idc=idc, sampleids=sampleids, patientids=patientids, pidc=pidc, files=files, verbose=verbose, verbose_all=verbose_all, like=like, missing=missing, order_by=order_by, top=top, by=by, print_query=print_query, vaa=vaa, jselect=jselect, lselect=lselect, joins=joins)
        # if raw is wished return the raw result.  39``
        if raw:
            return res
        # turn the result to samples.
        # 
        # the first implementation tried to take the tablenames/colums as keys
        # for the returned json fields, but that doesn't quite work for
        # patientpsn, where the key would be just 'idcontainer.psn'. so for now
        # joined-in fields are renamed to 'sampleid', 'patientid',
        # 'parentid' etc. // todo still true?  40``/sample/_
        sarr = []
        for r in res:
            # collect the sample's ids.  41.``
            ids = []
            #if dig(r, sampleid) is not None:
            #    ids.append(Identifier(id=dig(r, sampleid), code=self.sidc()))
            for idc in self._sidcs():
                #print("idc:" + idc)
                #print("r:" + str(r))
                if idc.lower() in r and r[idc.lower()] is not None:
                    ids.append( Identifier(id=dig(r, idc.lower()), code=idc.upper()) )
            # include the oid.  42``
            ids.append( Identifier(id=int(dig(r, sampleoid)), code="oid") ) # todo rename sampleoid?
            # reference the parent as an Idable to be able to include both oid and
            # sampleid. the oid is needed for aliquotgroups that don't come with an
            # idcontainer.  43``
            pids = []
            if r[parentoid] is not None:
                pids.append(Identifier(id=int(r[parentoid]), code="oid"))
            if dig(r, parentid) is not None:
                pids.append(Identifier(id=dig(r, parentid), code=self.sidc()))
            parent = None
            if len(pids) > 0:
                parent = Idable(ids=pids, mainidc=self.sidc())
            # also reference the patient as Idable.  44``
            patids = []
            for idc in self._pidcs(pidc):
                if idc.lower() in r and r[idc.lower()] is not None:
                    patids.append( Identifier(id=dig(r, idc.lower()), code=idc.upper()) )

            patient = None
            if len(patids) > 0:
                patient = Idable(ids=patids, mainidc=self.pidc(pidc))
            # build the sample.  45``
            s = Sample(
                appointment=dig(r, appointment),
                category=dig(r, category),
                change_date=dig(r, "change_date"), # change_*: for audit queries
                change_id=dig(r, "change_id"),
                change_kind=dig(r, "change_kind"),
                change_user=dig(r, "change_user"),
                concentration=dig(r, concentration),
                creationdate=dig(r, creationdate),
                cxxkitid=dig(r, cxxkitid),                
                derivaldate=dig(r, derivaldate),
                ids=Idable(ids=ids, mainidc=self.sidc()),
                initialamount=Amount(floatornull(dig(r, initialamount)), dig(r, initialunit)), # apparently the cast to float is explicitly needed
                kitid=dig(r, kitid),
                locationpath=dig(r, locationpath),
                locationname=dig(r, locationname),
                orga=dig(r, orga),
                parent=parent,
                patient=patient,
                project=dig(r, project),
                receiptdate=dig(r, receiptdate),
                receptacle=dig(r, receptacle),
                repositiondate=dig(r, repositiondate),
                restamount=Amount(floatornull(dig(r, restamount)), dig(r, restunit)),
                samplingdate=dig(r, samplingdate),
                secondprocessing=dig(r, secondprocessing),
                secondprocessingdate=dig(r, secondprocessingdate),
                stockprocessing=dig(r, stockprocessing),
                stockprocessingdate=dig(r, stockprocessingdate),
                trial=dig(r, trial),
                type=dig(r, type), 
                xposition=dig(r, xposition), 
                yposition=dig(r, yposition)
            )
            # buffer each fetch to r with dig, for keys that are not in r.
            # 
            # 
            # 
            # if by was passed, put the sample in the bydict by the specified key.  46``
            if by is not None:
                _fillby(bydict, by, r, s)
            # if by wasn't passed, append sample to the array.  47``
            else:
                sarr.append(s)
            #``41
        # return, depending on whether by was passed.  48``/sample/_
        if by is not None:
            if missing:
                return missinglst
            else:
                return bydict
        else:
            return sarr
        #``13
    #``12
    # _get_parents collects the parent sample oids of a sample in the `out`
    # list, in order from root to leaf. take oids to include aliquotgroups,
    # they don't necessarily have an idcontainer attached to them.  51``/_get_parents:
    def _get_parents(self, oid, out):
        # get the parent oid. query directly for performance (sample() would
        # join in the sampleid idcontainer by default).  52.``
        res = self.db.qfad("select parent from centraxx_sample where oid = ?", oid)
        pid = dig(res[0], "parent")
        # insert the parent at the front of the out array, to keep the root-to-leaf
        # direction.  53``
        if pid is not None:
            out.insert(0, pid)
            # recurse to get the parent's parent.  54.``
            self._get_parents(pid, out)
            #``54
        #``52
    #``51
    # _get_childs collects the oids of the sample's children in the `out`
    # list, in order from root-to-leaf. take oids to include aliquotgroups.  55``/_get_childs:
    def _get_childs(self, oid, out):
        # get the children. query directly for performancs (sample() would join
        # in the main sample idcontainer by default).  56.``
        res = self.db.qfad("select oid from centraxx_sample where parent = ?", oid)
        # if there are children, append them.  57``
        for child in res:
            out.append(child["oid"])
        # returse to get the children of the children.  58``
            self._get_childs(child["oid"], out)
        #``56
    #``55
    # patient gets patients and returns them as a list of Patient instances.
    # 
    # the parameters work analog to the sample method.  59``/patient: #py
    def patient(self, patientids:list|None=None, pidc:str|None=None, sampleids:list|None=None, idc:dict|None=None, trials:list|None=None, orgas:list|None=None, files:dict|None=None, verbose:list|None=None, verbose_all:bool=False, like:list|None=None, order_by:str|None=None, top:int|None=None, by:str=None, missing:bool=False, print_query:bool=False, raw:bool=False):
        # init arrays to avoid mutable default parameters.  60``_:
        if verbose is None:
            verbose = []
        if like is None:
            like = []
        if files is None:
            files = {}
        if idc is None:
            idc = {}
        # vaa (verbose-all-array) holds all possible verbose options for patient.  61`` #py
        vaa = [orga]
        # append the idcs for patient to the vaa array.  don't append idcs for
        # sample, cause then one patient with fifteen samples would come out as
        # fifteen result rows?  62``
        vaa.extend(self._pidcs(pidc))        
        # always join the patientid via verbose. put it as the first array element
        # for subsequent joins that might need it.  63`` #py
        if self.pidc(pidc) not in verbose:
            verbose.insert(0, self.pidc(pidc))
        # put the arguments in a dict of lists.  64``
        lists = {
          trial: trials,
          orga: orgas
        }
        # the select statements for optional joins by key.  65`` #py
        jselects = {
            self.pidc(pidc): [f"idc_{self.pidc(pidc)}.psn as '{patientid}'"],
            orga: [f"organisationunit.code as '{orga}'"],
            trial: [f"flexistudy.code as '{trial}'"],
        }
        # the selects for patientcontainer.  66``
        lselects = ["patientcontainer.*"]
        # the join statements for optional joins by key. if the same join is
        # needed for two different keys they are filtered out later.
        # 
        # pull the values from the join dict, jd.  67`` #py
        joins = {
            orga: self.jd["patient_to_orga"],
            trial: self.jd["patient_to_trial"]
        }
        # for any sample idc, join in sample, which is needed by
        # _join_idc.  68``
        for sidc in self._sidcs():
            joins[sidc] = self.jd["patient_to_sample"]
        # build the query and return the results.  69``
        (res, bydict, missinglst, by) = self._query(tablename="patientcontainer", lists=lists, sampleids=sampleids, patientids=patientids, pidc=pidc, files=files, verbose=verbose, verbose_all=verbose_all, like=like, order_by=order_by, top=top, by=by, missing=missing, print_query=print_query, vaa=vaa, jselect=jselect, lselect=lselect, joins=joins)
        # if raw is wished return the raw result.  70``
        if raw:
            return res
        # make patients from the result.  71``
        pats = []
        for r in res:
            # collect the ids.  72.``
            ids = []
            for idc in self._pidcs(pidc):
                if idc.lower() in r and r[idc.lower()] is not None:
                    ids.append( Identifier(id=dig(r, idc.lower()), code=idc) )
            # build the patient.  73``
            pat = Patient(
              ids=Idable(ids=ids, mainidc=self.pidc(pidc)),
              orga=dig(r, orga)
            )
            # if by was passed, put the pat in the bydict by the specified key.
            # 
            # if missing was passed, just delete the by-key from the missing list.  74``
            if by is not None:
                _fillby(bydict, by, r, pat)
            # if by wasn't passed, append sample to the array.  75``
            else:
                pats.append(pat)
            #``72
        # return, depending on whether by was passed.  76``/patient/_
        if by is not None:
            if missing:
                return missinglst
            else:
                return bydict
        else:
            return pats
        #``60
    #``59
    # trial gives trials.  79``/trial: #py
    def trial(self):
        # return the trial codes for now.  80.``
        query = "select code, study_name as name from centraxx_flexistudy"
        res = self.db.qfad(query)
        return res
        #``80
    #``79
    # finding gets the laborfindings ("messbefund" / "begleitschein") for
    # sampleids or method.  it returns a list of Finding instances.
    # 
    # you can pass these to verbose: tr.patientid  // maybe also tr.values?  81``/finding: #py
    def finding(self, sampleids:list|None=None, patientids:list|None=None, pidc:str|None=None, idc:dict|None=None, methods:list|None=None, trials:list|None=None, values:bool=True, verbose:list|None=None, files:dict|None=None, verbose_all:bool=False, names:bool=False, top:int|None=None, print_query:bool=False, raw:bool=False):
        # init arrays to avoid multiple default parameters.  82``/finding/init:
        if verbose is None:
            verbose = []
        if files is None:
            files = {}
        if idc is None:
            idc = {}
        #``82
        # vaa (verbose-all-array) holds all possible verbose options for patient.  83``/finding/findings: #py
        vaa = [patientid] # include trial?
        # on verbose_all include all constants in the verbose array.  84`` #py
        if verbose_all:
            verbose = vaa
        # for now always include the sampleid.  85``
        if self.sidc() not in verbose:
            verbose.append(self.sidc())
        # replace traction constants patientid and sampleid with their idc.  86``
        verbose = self._concrete_idcs(verbose, pidc=pidc)
        files = self._concrete_idcs_dict(files, pidc=pidc)
        # put the arguments in a dict of lists.  87``
        lists = {
          method: methods,
          trial: trials
        }
        # resolve the sampleid and patientid shortcut by putting their lists
        # into the idc dict.  88``
        _dextend(idc, self.sidc(), sampleids)
        _dextend(idc, self.pidc(pidc), patientids)
        # now idc holds all parameters queryable as idcontainers.
        # 
        # read the lists from files and put them together with the lists passed.  89``
        alllists = self._makealllists(lists, idc, files)
        # put all lists longer than 50 items into temporary tables.
        # 
        # nontable holds the remaining lists, table holds the names of the created
        # temporary tables, each distinguishing nonidc and idc.  90``
        (nontable, table) = self._makemove(alllists, 50)
        # build the query for findings.  later add recorded values to each finding.
        # 
        # get the idc part of the select string.  91``
        selects = {
            self.sidc(): [f"idc_{self.sidc()}.psn as '{sampleid}'"],
            self.pidc(pidc): [f"idc_{self.pidc(pidc)}.psn as '{patientid}'"],                   
        }
        idcselectstr = self._selectstr(selects, [], nontable, table, verbose, pidc=pidc)  
        # for any patient idc, join in patientcontainer, which is needed by
        # _join_idc.  92``
        joins = {
            trial: self.jd["sample_to_trial"]
        }
        for pidc in self._pidcs(pidc):
            joins[pidc] = self.jd["sample_to_patient"]
        # get the join string.
        # 
        # todo why called idcjoinstr?  93``
        idcjoinstr = self._joinstr(joins, nontable, table, verbose, pidc=pidc)
        # get the top string.  94``
        topstr = self._top(top)
        # stick together the query.  95``
        query = f"""select {topstr} laborfinding.oid as "laborfinding_oid", laborfinding.*, labormethod.code as {method}, {idcselectstr}
        from centraxx_laborfinding as laborfinding

        -- go from laborfinding to sample
        left join centraxx_labormethod as labormethod on laborfinding.labormethod = labormethod.oid
        left join centraxx_labormapping as labormapping on labormapping.laborfinding = laborfinding.oid
        left join centraxx_sample sample on labormapping.relatedoid = sample.oid
        {idcjoinstr}"""
        # get the where string and append it to the query.  96``
        (wherestr, whereargs) = self._where(nontable, table)
        if wherestr.strip() != "":
            query += f"\nwhere {wherestr}"
            
        if print_query:
            print(query)
            print(whereargs)
        # get the result.  97``
        results = self.db.qfad(query, whereargs)
        # remove the temporary tables holding bcp data.  98``
        self._cleartt(table["nonidc"])
        self._cleartt(table["idc"])        
        #``83
        # for each of the findings, pull in the recorded values.
        # 
        # maybe only on --verbose = ["values"]
        # 
        # if names should be loaded, get the laborvalue display names. load them
        # once to avoid joining them in each time, is that quicker?
        # 
        # catalogentry and usageentry display names are loaded in _make_rec if
        # needed.  99``/finding/values:
        if names is True:
            self.names_laborvalue = self.name(table="laborvalue")
        # now iterate the findings and get their recorded values.  100``
        for i, finding in enumerate(results):
            if not values: # todo put this outside of the loop?
                continue
            # construct the query for the recorded values.
            # 
            # go from laborfinding to recordedvalue via labfindinglabval.
            # 
            # todo prefetch units?  101``/finding/values/query:
            query = """select recordedvalue.*, laborvalue.code as laborvalue_code, laborvalue.dtype as laborvalue_type, laborvalue.custom_catalog as laborvalue_catalog_oid, unit.code as laborvalue_unit
                from centraxx_laborfinding as laborfinding

                -- go from laborfinding to recorded value
                join centraxx_labfindinglabval as labfindinglabval on labfindinglabval.laborfinding = laborfinding.oid
                join centraxx_recordedvalue as recordedvalue on labfindinglabval.oid = recordedvalue.oid"""
            # join the laborvalue (messparameter) differently in cxx 3 and cxx 4.  102``
            if self.cxx() == "3":
                query += """--go directly to laborvalue
                join centraxx_laborvalue laborvalue on labfindinglabval.laborvalue = laborvalue.oid"""
            elif self.cxx() == "4":
                query += """                
                -- go to laborvalue via crftempfield
                join centraxx_crftempfield crftempfield on labfindinglabval.crftempfield = crftempfield.oid
                join centraxx_laborvalue laborvalue on crftempfield.laborvalue = laborvalue.oid
                """
            # continue the query the same.  103``
            query += """
                --go from laborvalue to unit
                left join centraxx_unity unit on laborvalue.unit = unit.oid

                where laborfinding.oid = ?
            """
            # turn the result into recorded value objects. key the values by
            # laborvalue_code (the messparam code).  104``
            recvals = self.db.qfad(query, finding['laborfinding_oid'])
            valsbycode = {}
            for recval in recvals:
              valsbycode[recval["laborvalue_code"]] = self._make_rec(recval, finding, names)
            #``101
            # put the values to the finding.  105``/finding/values/put:
            results[i]["values"] = valsbycode
            #``105
        #``99
        # if raw, return.  106``/finding/ret:
        if raw:
            return results
        # make findings from the results.  107``
        findings = []
        for res in results:
            # make the patient idable.  108.``
            patids = []
            for idc in self._pidcs(pidc):
                if idc.lower() in res and res[idc.lower()] is not None:
                    patids.append( Identifier(id=dig(res, idc.lower()), code=idc.upper()) )

            patient = None
            if len(patids) > 0:
                patient = Idable(ids=patids, mainidc=self.pidc(pidc))
            # make the finding.  109``
            finding = Finding(
                findingdate=dig(res, "findingdate"),
                method=res["method"],
                name=res["shortname"],
                patient=patient,
                recs=res["values"] if "values" in res else None, # todo None ok?
                sample=Idable(id=res[sampleid], code=self.sidc(), mainidc=self.sidc()),
                sender=None
            )
            findings.append(finding)
            #``108
        # return the findings with their respective values.  110``/finding/ret
        return findings
        #``106
    #``81
    # method (messprofil) gets method(s) and their labvals (messparameter).  121``/method:
    def method(self, methods:list|None=None, files:dict|None=None):
        # init files.  122.``
        if files is None:
            files = {}
        # put the arguments in a dict of lists.  123``
        lists = {
          method: methods
        }
        # now idc holds all parameters queryable as idcontainers.
        # 
        # read the lists from files and put them together with the lists passed.  124``
        alllists = self._makealllists(lists, {}, files)
        # put all lists longer than 50 items into temporary tables.
        # 
        # nontable holds the remaining lists, table holds the names of the created
        # temporary tables, each distinguishing nonidc and idc.  125``
        (nontable, table) = self._makemove(alllists, 50)
        # the query. get from labormethod to laborvalue via the crf tables.  126``
        query = """select laborvalue.code as labval, labormethod.code as "method", laborvalue.dtype as labval_type, catalog.code as catalog
from centraxx_labormethod labormethod
inner join centraxx_crftemplate crf_t
    on labormethod.crf_template=crf_t.oid
inner join centraxx_crftempsection crf_ts
    on crf_t.oid=crf_ts.crftemplate
inner join centraxx_crftempsection_fields crf_tsf
    on crf_ts.oid=crf_tsf.crftempsection_oid
inner join centraxx_crftempfield crf_tf
    on crf_tsf.crftempfield_oid=crf_tf.oid
inner join centraxx_laborvalue laborvalue
    on crf_tf.laborvalue=laborvalue.oid
left join centraxx_catalog catalog
    on catalog.oid = laborvalue.custom_catalog"""
        # add the where string.  127``
        (wherestr, whereargs) = self._where(nontable, table)

        if wherestr and wherestr.strip != "()":
          query += " where " + wherestr
        # print(query)
        # return the result.  128``
        res = self.db.qfad(query, whereargs)
        # remove the temporary tables.  129``
        self._cleartt(table["nonidc"])
        self._cleartt(table["idc"])
        # if raw, return result as is.
        # 
        # `
        # if raw:
        # return res
        # `
        # 
        # get the names.  todo cache?  130``
        methodnames = self.name(table="labormethod")
        labvalnames = self.name(table="laborvalue")
        # key the labvals by method code, add the names.
        # 
        # todo why are some labvals returned more than one in the query?  131``
        out = {}
        for row in res:
            # add the method.  132.``
            methodcode = row["method"]
            if methodcode not in out:
                out[methodcode] = {}
                out[methodcode]["code"] = methodcode
                out[methodcode]["name_de"] = dig(methodnames, methodcode + "/de")
                out[methodcode]["name_en"] = dig(methodnames, methodcode + "/en")
                out[methodcode]["labvals"] = {}
            # add labvals to method.  133``
            labvalcode = row["labval"]

            labval = {}
            labval["code"] = labvalcode
            labval["name_de"] = dig(labvalnames, labvalcode + "/de")
            labval["name_en"] = dig(labvalnames, labvalcode + "/en")
            if dig(row, "catalog") is not None:
                labval["catalog"] = dig(row, "catalog")
            labval["type"] = row["labval_type"]
            # for optiongroups and enumerations, add the allowed controlled vocabulary for this labval.  134``
            if labval["type"] == "OPTIONGROUP" or labval["type"] == "ENUMERATION":
                labval["usageentry"] = self.usageentry(labvals=[labval["code"]])
            # add the labval.  135``
            out[methodcode]["labvals"][labvalcode] = labval
            # return.  136``
        return out
            #``132
        #``122
    #``121
    #   137``/user:
    def user(self, usernames:list|None=None, emails:list|None=None, lastlogins:list|None=None, files:dict|None=None, verbose:list|None=None, top:int|None=None, verbose_all:bool=False, print_query:bool=False):
        # init arrays to avoid multiple default parameters.  138.``
        if verbose is None:
            verbose = []
        if files is None:
            files = {}
        # verbose all array holds the possible verbose options.
        # 
        # address and login are catch-alls for address and login related fields.  139``
        vaa = [address, login]
        # for now.   140``
        if verbose_all:
            verbose = vaa
        # put the arguments in a dict of lists.  141``
        lists = {
          username: usernames,
          address: emails,
          login: lastlogins
        }
        # now idc holds all parameters queryable as idcontainers.
        # 
        # 
        # read the lists from files and put them together with the lists passed.  142``
        alllists = self._makealllists(lists, {}, files)
        # put all lists longer than 50 items into temporary tables.
        # 
        # nontable holds the remaining lists, table holds the names of the created
        # temporary tables, each distinguishing nonidc and idc.  143``
        (nontable, table) = self._makemove(alllists, 50)
        # jselects holds the joined in selects.  144``
        jselects = {
            address: [f"address.*"],
            login: [f"credential.last_login_on as lastlogin"]
        }
        # lselects holds the local selects for this table, to rename fields.
        # 
        # put sample.* first, that it doesn't overwrite field of the renames?  145``
        lselects = [
            f"participant.*"
        ]
        # put together the select string.  146``
        selectstr = self._selectstr(jselects, lselects, nontable, table, verbose)
        # joins holds the join statements needed.  147`` #py
        joins = {
            address: self.jd["participant_to_address"],
            login: self.jd["participant_to_credential"]
        }
        # put together the join string for the query.  148`` #py
        joinstr = self._joinstr(joins, nontable, table, verbose)
        # get the where string.  149`` #py
        (wherestr, whereargs) = self._where(nontable, table, like=[])
        # get the top string.  150``
        topstr = self._top(top)
        # stick together the query.  151`` #py
        query = f"select {topstr} {selectstr} from centraxx_participant participant\n{joinstr}"
        if wherestr.strip() != "":
            query += f"\nwhere {wherestr}"
        # print if wished.  152``
        if print_query:
            print(query)
            print(whereargs)
        # run the query.  153``
        res = self.db.qfad(query, whereargs)
        # remove the temporary tables holding bcp data.  154``
        self._cleartt(table["nonidc"])
        self._cleartt(table["idc"])        
        # turn the result into user instances.  155``
        out = []
        for r in res:
            my_address = Address(
                    country_descriptor = dig(r, "country_descriptor"),
                    city = dig(r, "city"),
                    email = dig(r, "email"),
                    fax = dig(r, "fax"),
                    mobile = dig(r, "mobile"),
                    phone1 = dig(r, "phone1"),
                    phone2 = dig(r, "phone2"),
                    street = dig(r, "street"),
                    street_no = dig(r, "streetno"),
                    zip_code = dig(r, "zipcode"),
                    po_box = dig(r, "po_box"),
                    contact_person = dig(r, "contact_person"),
                    contact = dig(r, "contact")
            )
            user = User(
                address= my_address,
                divisional_admin=dig(r, "divisional_admin"),
                lastlogin=dig(r, lastlogin),
                username=dig(r, username),
                firstname=dig(r, "firstname"),
                lastname=dig(r, "lastname")
            )
            out.append(user)
        # return.  156``
        return out
        #``138
    #``137
    # orga returns organisationunits.  159``/orga:
    def orga(self, names:bool=False):
        # make the query.  160``_:
        query = "select code from centraxx_organisationunit"
        # run the query.  161``
        res = self.db.qfad(query)
        # if names should be included, include them.  162``
        names = self.name(table="organisationunit", ml_table="organisatunit_ml_name")
        for i, _ in enumerate(res):
            nam = names[res[i]["code"]]
            res[i]["name_de"] = nam["de"]
            res[i]["name_en"] = nam["en"]
        # return.  163``
        return res
        #``160
    #``159
    # catalog gives the catalogentries per catalog.  164``/catalog:
    def catalog(self, catalogs:list|None=None, files:dict|None=None):
        # initialize None params.  165.``
        if files is None:
            files = {}
        # put the arguments in a dict of lists.  166``
        lists = {
            catalog: catalogs
        }
        # read the lists from files and put them together with the lists passed.  167``
        alllists = self._makealllists(lists, {}, files)
        # put all lists longer than 50 items into temporary tables.
        # 
        # nontable holds the remaining lists, table holds the names of the created
        # temporary tables, each distinguishing nonidc and idc.  168``
        (nontable, table) = self._makemove(alllists, 50)
        # query the catalogentries, joining in their cataloges.  169``
        query = """select catalogentry.code as 'entry_code', catalog.code as 'catalog_code' from centraxx_catalogentry catalogentry
join centraxx_catalog catalog on catalogentry.catalog = catalog.oid"""
        # make the wherestring.  170``
        (wherestr, whereargs) = self._where(nontable, table) 
        if wherestr and wherestr.strip() != "()":
            query += f"\nwhere {wherestr}"
        # get the result.  171``
        res = self.db.qfad(query, whereargs)
        # remove the temporary tables.  172``
        self._cleartt(table["nonidc"])
        self._cleartt(table["idc"])
        # get the names for catalog entries and cataloges.  173``
        catnames = self.name(table="catalog")
        entrynames = self.name(table="catalogentry")
        # put the output together. key by catalog code. add the display names.  174``
        out = {}
        for row in res:
            entrycode = dig(row, "entry_code")
            catcode = dig(row, "catalog_code")
            # create the catalog
            if catcode not in out:
                out[catcode] = {}
                out[catcode]["code"] = catcode
                out[catcode]["name_de"] = dig(catnames, catcode + "/de")
                out[catcode]["name_en"] = dig(catnames, catcode + "/en")
                out[catcode]["entries"] = {}
            # add the entry
            entry = {}
            entry["code"] = entrycode
            entry["name_de"] = dig(entrynames, entrycode + "/de")
            entry["name_en"] = dig(entrynames, entrycode + "/en")        
            out[catcode]["entries"][entrycode] = entry
        # return.  175``
        return out
        #``165
    #``164
    # usageentry gives the usageentries.  176``/usageentry:
    def usageentry(self, labvals:list=None, files:dict=None):
        # put the arguments in a dict of lists.  177.``
        lists = {
           labval: labvals
        }
        # read the lists from files and put them together with the lists passed.  178``
        alllists = self._makealllists(lists, {}, files)
        # put all lists longer than 50 items into temporary tables.  179``
        (nontable, table) = self._makemove(alllists, 50)
        # put together the select string.  180``
        selectstr = self._selectstr([], ["usageentry.code"], nontable, table, verbose=[])
        # joins holds the join statements needed.  181`` #py
        joins = {
                labval: self.jd["usageentry_to_labval"]
        }
        # put together the join string for the query.  182`` #py
        joinstr = self._joinstr(joins, nontable, table, [])
        # get the where string.  183`` #py
        (wherestr, whereargs) = self._where(nontable, table, like=[])
        # stick together the query, get the result and return it.  184`` #py
        query = f"select {selectstr} from centraxx_usageentry usageentry \n{joinstr}"
        if wherestr.strip() != "":
            query += f"\nwhere {wherestr}"
        # if print_query:
        # print(query)
        # print(whereargs)
        # get the result.  185``
        res = self.db.qfad(query, whereargs)
        # clear the temporary tables.  186``
        self._cleartt(table["nonidc"])
        self._cleartt(table["idc"])
        # get the display names (todo cache)?  187``
        names = self.name(table="usageentry")
        # put the names to the codes.  188``
        out = {}
        for row in res:
            code = row["code"]
            out[code] = {}
            out[code]["name_de"] = dig(names, code + "/de")
            out[code]["name_en"] = dig(names, code + "/en")
            out[code]["code"] = code
        # return.  189``
        return out
        #``177
    #``176
    # name gives the multilingual names for a code or all codes in a table.
    # 
    # the result is keyed by code and language like this:
    # 
    # 
    # "NUM_NMR_ISOLEUCINE_VALUE": {  
    # "de": "Isoleucin",  
    # "en": "Isoleucine"  
    # }
    # 
    # 
    # table: the name of the centraxx table without centraxx_ prefix  
    # code: a specific code, if none given, all code - name mappings for table are given  
    # lang: de|en  
    # ml_table (cxx3): if the name of the table connecting to multilingualentry is not simlpy the queried table name followed by "_ml_name", give the connecting table's name here. eg: name('laborvaluegroup', ..., ml_name='labval_grp_ml_name')  191``/name:
    def name(self, table:str, code:str=None, lang:str=None, ml_table:str=None):
        # the db queries differ for cxx3 and cxx4.  192.``
        if self.cxx() == "3":
            return self._name_cxx3(table, code, lang, ml_table)
        elif self.cxx() == "4":
            return self._name_cxx4(table, code, lang)
        #``192
    #``191
    # _name_cxx3 gives display names for cxx 3.  193``/_name_cxx3:
    def _name_cxx3(self, table:str, code:str=None, lang:str=None, ml_table:str=None):
        # check that table and ml_table are valid sql identifiers.  194.``
        if not isidentifier(table) or (ml_table is not None and not isidentifier(ml_table)):
            raise Exception(f"names {table} and {ml_table} need to be valid sql identifiers.")
        # interlacing the table name assumes that the referencing pattern for
        # multilingual entries stays the same across table names.  195``
        query = "select [" + table + "].code, multilingual.value as name, multilingual.lang as lang"
        query += " from [centraxx_" + table + "] as [" + table + "]"
        # put together the name for the ml_table.  196``
        ml_name = ""
        if ml_table is not None: # the name is different
            ml_name = "centraxx_" + ml_table
        else: # the name is the same
            ml_name = "centraxx_" + table + "_ml_name"
        # add it to the query.  197``
        query += " inner join [" + ml_name + "] mlname on mlname.related_oid = [" + table + "].oid"
        query += " inner join centraxx_multilingualentry multilingual on mlname.oid = multilingual.oid"
        # restrict the query to specific lang or code if given.  198``
        wherestrings = []
        args = []
        if code is not None:
            wherestrings.append(self._whereparam("[" + table + "].code"))
            args.append(code)
        if lang is not None:
            wherestrings.append(self._whereparam("multilingual.lang"))
            args.append(lang)
        # only add sql-where if needed.  199``
        if len(wherestrings) > 0:
            query += " where "
            # join where clauses by and
            query += " and ".join(wherestrings)

        # print(query)
        # query.  200``
        res = self.db.qfad(query, *args)
        # structure by code and lang and return.  201``
        out = {}
        for line in res:
            code = line["code"]
            lang = line["lang"]
            if code not in out:
               out[code] = {}
            out[code][lang] = line["name"]
        return out
        #``194
    #``193
    # _name_cxx4 gives display names for cxx 4.  202``/_name_cxx4:
    def _name_cxx4(self, table:str, code:str=None, lang:str=None):
        # check that table is a valid sql identifier.  203.``
        if not isidentifier(table):
            raise Exception(f"name {table} needs to be a valid sql identifier.")
        # the field name is exactly as the table name, except for organisationunit.  204``
        fieldname = table
        if table == "organisationunit":
            fieldname = "organizationunit"
        # the query.  205``
        query = f"select lang, ml_name as name, {table}.code from centraxx_multilingual multilingual join centraxx_{table} {table} on {table}.oid = multilingual.{table} where multilingual.{fieldname} is not null"
        # the following is actually the same as for cxx 3, scrape the seperate
        # _name_cxx3 and _name_cxx4 methods?
        # 
        # restrict the query to specific lang or code if given.  206``
        wherestrings = []
        args = []
        if code is not None:
            wherestrings.append(self._whereparam("[" + table + "].code"))
            args.append(code)
        if lang is not None:
            wherestrings.append(self._whereparam("multilingual.lang"))
            args.append(lang)
        # only add sql-where if needed.  207``
        if len(wherestrings) > 0:
            query += " where "
            # join where clauses by and
            query += " and ".join(wherestrings)
        # query.  208``
        res = self.db.qfad(query, *args)
        # structure by code and lang and return.  209``
        out = {}
        for line in res:
            code = line["code"]
            lang = line["lang"]
            if code not in out:
               out[code] = {}
            out[code][lang] = line["name"]
        return out
        #``203
    #``202

    # query
    # _selectstr gives the selects for idc keys and verbose array. unlike
    # _joinstr, keys in lists and tmptables aren't included automatically,
    # cause they may be needed in where but are not necessarily interesting
    # for the output (is that true?). selecta is for fields that should be
    # selected regardless if they're in the verbose array or not.
    # 
    # the idc argument assumes that the sample table is joined it. // todo is this still true?  234``/_selectstr: #py
    def _selectstr(self, selects, selecta, nontable, table, verbose, pidc:str=None):
        # split verbose into non-idc and idc.
        # 
        # collect the nonidc/idc keys from non tables and tables, respectively.  235.``
        (vnonidc, vidc) = self._splitidc(verbose, pidc=pidc)
        nonidc = self._collkeys(nontable["nonidc"], table["nonidc"], vnonidc)
        idc = self._collkeys(nontable["idc"], table["idc"], vidc)
        # put in the selects for non-idc.  236``
        for key in nonidc:
            # skip keys that are not in the selects dict (this way the calling
            # functions can control which keys not to include in the result, even
            # though they might be needed in the where string).  237.``
            if key not in selects:
                continue
            # put in the select line(s).  238``
            for s in selects[key]:
                selecta.append(s)
            #``237
        #``235
    # append idc selects.  239``/_selectstr #py
        selecta = self._select_idc(selecta, idc)
    # get the selections and joins as string.  240`` #py
        selectstr = ", \n".join(selecta)
        return selectstr
    #``234
    # _select_idc adds the sql select statements for a list of idc keys.  241``/_select_idc:
    def _select_idc(self, selecta, idca:list):
        # make select strings and append them.  242.``
        for item in idca:
          selectstr = f"idc_{item}.psn as '{item}'"
          if selectstr not in selecta:
            selecta.append(selectstr)
        # return the selecta with appended strings.  243``
        return selecta
        #``242
    #``241
    # _joinstr puts together the joins needed for the lists and temporary
    # tables of non-idc and idc keys, and verbose. it returns the sql join
    # string.  244``/_joinstr: #py
    def _joinstr(self, joins, nontable, table, verbose, pidc:str|None=None):
        # make two lists of what to join in, one for non-idc, one for idc.  245.``
        (vnonidc, vidc) = self._splitidc(verbose, pidc=pidc)
        nonidc = self._collkeys(nontable["nonidc"], table["nonidc"], vnonidc)
        idc = self._collkeys(nontable["idc"], table["idc"], vidc)
        # first put in the idc joins, cause other joins might need them.  246``
        joina = []
        joina = self._join_idc(joina, idc, joins)
        # now put in joins for non-idc.
        # 
        # could you substract idc from verbose here cause there were already taken care of?  247``
        for key in nonidc: 
            # skip keys that are not in the joins dict // still necessary if idc-joins are sorted out before?  248.``
            if key not in joins:
                continue
            # put in the join line(s). make sure that join clauses aren't included
            # double, for example both locationpath and locationname join in
            # samplelocation, so check that it's not already in the join array.  249``
            for s in joins[key]:
                if s not in joina:
                    joina.append(s)
            # get the joins as string.  250`` #py
        joinstr = " \n".join(joina)
        return joinstr
            #``248
        #``245
    #``244
    # _join_idc adds the sql join statements for the idc keys of the query.  251``/_join_idc: #py
    def _join_idc(self, joina, idca:list, joins):
        # add a join for each member of the idc array.
        # 
        # the joins are different for idckind SAMPLE and PATIENT.  252.``
        for item in idca:
          # put in idc-intermediary joins, for when sample joins in patient idcs and
          # needs patientcontainer for that or vice versa.   253``./intermediary:
          if item in joins:
            for s in joins[item]:
              if s not in joina:
                joina.append(s)
          #``253
          if self._idckind[item] == "SAMPLE":
            # join in sampleidcontainer for SAMPLE. prefix it with idc_ instead of
            # sidc_, so that there can be one where-check that uses the idc_ prefix
            # for both patient and sample idcontainers.
            # 
            # when joining in idcontainers, we need to check their respective
            # idcontainer type. do this here as opposed to the where clause at the
            # end of the query, cause here if a idcontainer is not set checking here
            # just makes the field null in the result row, whereas checking in the
            # where clause it excludes the entire row.  254``../sample:
            joinstr = f"left join centraxx_sampleidcontainer as idc_{item} on idc_{item}.sample = sample.oid and idc_{item}.idcontainertype = {self._idcoid[item]}"
    
            if joinstr not in joina:
              joina.append(joinstr)
            #``254
          elif self._idckind[item] == "PATIENT":
            # join in idcontainer via patientcontainer for PATIENT:  255``../patient:
            joinstr = f"left join centraxx_idcontainer as idc_{item} on idc_{item}.patientcontainer = patientcontainer.oid and idc_{item}.idcontainertype = {self._idcoid[item]}"
            if joinstr not in joina: # neccessary?
              joina.append(joinstr)
            #``255
          else:
            print(f"error: idcontainer kind {self._idckind[item]} not supported.")
        #``252
    # return the selecta with appended strings.  256``/_join_idc
        return joina
    #``251
    # _where returns the wherestring and args array for the provided
    # lists and temporary tables.
    # 
    # the user needs to make sure that whatever is used here is joined
    # into the query before.
    # 
    # the keys in nontable["nonidc"] and table["nonidc"] are assumed to be
    # the tr constants, e.g. tr.locationpath, the keys in nontable["idc"]
    # and table["idc"] are assumed to be idcontainers.
    # 
    # like holds tr constants and idcontainers for which to check likeness
    # instead of equality.  257``/_where:
    def _where(self, nontable:dict, table:dict, like:list=None): # -> (str, list)
        # init arrays to avoid mutable default parameters.  258.``
        if like is None:
            like = []
        # map which tr constant point to which db table and column.
        # 
        # for now mark dates with type date.  259``
        wheredict = { 
          trial: { "field": "flexistudy.code" },
          locationpath: { "field": "samplelocation.locationpath" },
          locationname: { "field": "samplelocation.locationid" },          
          method: { "field": "labormethod.code" },
          kitid: { "field": "samplekit.kitid" },
          cxxkitid: { "field": "samplekit.cxxkitid" },
          catalog: { "field": "catalog.code" },
          category: { "field": "sample.dtype" },
          type: { "field": "sampletype.code" },          
          orga: { "field": "organisationunit.code" },
          parentid: { "field": "parentidc.psn" },
          parentoid: { "field": "sample.parent" },
          sampleoid: { "field": "sample.oid" },          
          samplingdate: { "field": "sample.samplingdate", "type": "date" },
          receiptdate: { "field": "sample.receiptdate", "type": "date" },
          derivaldate: { "field": "sample.derivaldate", "type": "date" },
          first_repositiondate: { "field": "sample.first_repositiondate", "type": "date" },
          repositiondate: { "field": "sample.repositiondate", "type": "date" },
          stockprocessingdate: { "field": "sample.stockprocessingdate", "type": "date" },
          secondprocessingdate: { "field": "sample.secondprocessingdate", "type": "date" },
          username: { "field": "participant.username" },
          address: { "field": "address.email" },
          login: { "field": "credential.last_login_on", "type": "date" },
          labval: { "field": "labval.code" }
        }
        # make arrays to hold where clauses and their respective arguments.  260``
        wherestrs = []
        whereargs = []
        # go over the non-idc and idc keys seperately to distinguish in the
        # probably rare case that an idcontainer should be named like a tr constant.
        # 
        # for now do dates only for lists, not tmp tables.
        # 
        # go over the non-idc keys.  261``
        for key in _uniq(list(nontable["nonidc"].keys()) + list(table["nonidc"].keys())):
           if dig(nontable["nonidc"], key) is not None and dig(wheredict, key + "/type") == "date":
               (wstr, wargs) = self._wheredate(dig(nontable["nonidc"], key), wheredict[key]["field"])
           else:
               (wstr, wargs) = self._whereexact(dig(nontable["nonidc"], key), dig(table["nonidc"], key), wheredict[key]["field"])
           if wstr != "()":
               wherestrs.append(wstr)
               whereargs.extend(wargs)
        # go over the idc keys.
        # 
        # don't check for date here.  262``
        for key in _uniq(list(nontable["idc"].keys()) + list(table["idc"].keys())):
            (wstr, wargs) = self._whereexact(dig(nontable["idc"], key), dig(table["idc"], key), f"idc_{key}.psn")
            if wstr != "()":
                wherestrs.append(wstr)
                whereargs.extend(wargs)
        # chain the where clauses by 'and'.  263``
        wherestr = " and ".join(wherestrs)
        # return the where string and its arguments.  264``
        return (wherestr, whereargs)
        #``258
    #``257
    # _whereexact returns a where clause and argument list for exact
    # matching of items in list and temporary table. NULL string is list is
    # treated as sql NULL value.  269``/_whereexact:
    def _whereexact(self, lst, tmptable, dbfield:str): # -> (str, list)
        # open the where string.  270.``
        wherearg = []
        wherestr = "("
        # make null checks extra, e.g. '(kitid is NULL or kitid in (1, 2, 3))'.
        # 
        # if is there a 'NULL' in the array, put in a NULL check and remove it
        # from the array, so it doesn't mess with the where-arguments later.  271``
        needsor = False
        if lst is not None and 'NULL' in lst:
            lst.remove("NULL")
            wherestr += dbfield + " is NULL"
            needsor = True
        # make the in-list part of the wherestring with the ?-placeholder
        # string.  272``
        if lst is not None and len(lst) > 0:
            if needsor:
                wherestr += " or "
            placeholder = traction._sqlinplaceholder(len(lst)) # todo put in package? tr._sqlinplaceholder
            wherestr += dbfield + " in " + placeholder # e.g. samplelocation.locationpath in (?, ?, ?)
            wherearg.extend(lst)
            # if something comes next, chain it with or
            needsor = True
        # check against the temporary table  273``
        if tmptable is not None:
            if needsor:
                wherestr += " or "
            wherestr += dbfield + " in "
            wherestr += f"(select stdin from tempdb..{tmptable})"
        # close the where string, return with the arg list.  274``
        wherestr += ")"
        return (wherestr, wherearg)
        #``270
    #``269
    # _wheredate makes a date check. for now only check that it's between
    # the two values of the array passed as date parameter.  265``/_wheredate:
    def _wheredate(self, tpl, dbfield): # -> (str, list)
        # for null checks, instead of the date tuple, 'NULL' is passed. maybe a
        # bit hacky?  266.``
        wstr = ""
        wargs = []
        if tpl == 'NULL':
            wstr = "(" + dbfield + " is NULL)"
        # check that a date is between the two dates in tpl.
        # 
        # CAST removes the time component, this is needed for getting samples
        # from one day (when both dates are the same).
        # 
        # ISNULL handles missing start or end date. is this smart, or rather
        # check for >=, <= etc if only one date is given?  267``
        else:
            s = "(CAST(" + dbfield + " as date) between ISNULL(?, " + dbfield + ") and ISNULL(?, " + dbfield + "))"
            wstr = s
            wargs.extend(tpl)
        # return where clause and args.  268``
        return (wstr, wargs)
        #``266
    #``265
    # _sqlinplaceholder returns a string like (?, ?, ?, ? ...) with n question marks for sql in.  275``/_sqlinplaceholder:
    def _sqlinplaceholder(n):

        # put this in a package sqlutil?

        out = "("
        for i in range(n):
            out += "?"
            if i < n - 1:
                out += ","
        out += ")"
        return out
    #``275
    # _whereparam gives a ?-parameterized sql where expression for name
    # equal or like parameter for use in queries.  276``/_whereparam:
    def _whereparam(self, name, like:bool=None):
        if like is None or not like:
            return name + " = ?"
        else:
            return name + " like '%' + ? + '%'"
    #``276
    # _wherelike gives a ?-parameterized sql where like expression.  277``/_wherelike:
    def _wherelike(self, name):
        return name + " like '%' + ? + '%'"

    #``277
    # _wherelikes gives a ?-parameterized sql of or-joined where-like
    # expressions, as many as given n.  278``/_wherelikes:
    def _wherelikes(self, field, n):
        a = []
        for i in range(n):
            a.append(field + " like '%' + ? + '%'") # the sql takes literal plusses like here
        return " or ".join(a)
    #``278
    # _top returns an injection-safe top string, for, e.g. `select top 100 * from table`, if
    # top is None return an empty string.  279``/_top:
    def _top(self, top):
        # exit if the top param isn't a number. better use a sql building library for that?  280.``
        if top is not None:
           if not isnumber(top):
               raise Exception(f"top param {top} needs to be a number")
           return f"top {top}"
        return ""
        #``280
    #``279
    # _order_by returns an injection safe order by string.  281``/_order_by:
    def _order_by(self, order_by):
        # exit if the order by parameter isn't an identifier.  282.``
        if not isidentifier(order_by):
            raise Exception(f"order_by param {order_by} needs to be an sql identifier.")
        return f"order by {order_by}"
        #``282
    #``281
    # _idcinit makes a mapping of idcontainers from their code (lower-case)
    # to respective oid and kind, and makes a list containing all idcs
    # including sidc and pidc.  283``/_idcinit:
    def _idcinit(self):
        # get the codes and oids.  284.``
        query = "select code, oid, kind from centraxx_idcontainertype"
        res = self.db.qfad(query)
        # return idcontainer oids and kinds keyed by their code.  285``
        self._idcoid = {}
        self._idckind = {}
        for row in res:
          self._idcoid[row["code"]] = row["oid"]
          self._idckind[row["code"]] = row["kind"]
        # make a list of all idcontainers that makes sure to include sidc and pidc.  286``
        self._idcs = []
        for id in self.settings['idc'] + [self.sidc(), self.pidc()]:
            if id not in self._idcs:
                self._idcs.append(id)
        #``284
    #``283
    # _make_rec makes a recorded value instance for finding from db results with display names if wished.  112``/_make_rec:
    def _make_rec(self, recval, finding, names:bool=False) -> Rec:
        # make different rec val instances depending on the type. see the Rec
        # type attribute for an overview of the different types.
        # 
        # convert the boolean values from integer to boolean.  113.``
        out:Rec = None
        if recval["laborvalue_type"] == "BOOLEAN":
            val = True if recval["boolvalue"] == 1 else False
            out = BooleanRec(method=finding["method"], labval=recval["laborvalue_code"], rec=val)
        # take DECIMAL values from the numericvalue field.
        # 
        # somehow the cast to float seems necessary, if the Decimal from the db
        # result is passed, the value ends up being None.
        # 
        # if a number value isn't set, the db returns None (and not 0.0 or so).  114``
        elif recval["laborvalue_type"] == "DECIMAL":
            #print(recval["laborvalue_code"])
            #print(recval)
            value = float(recval["numericvalue"]) if recval["numericvalue"] is not None else None
            out = NumberRec(method=finding["method"], labval=recval["laborvalue_code"], rec=value, unit=recval["laborvalue_unit"])
        # there are only four laborvalues for INTEGER, and they don't seem to be
        # findable with erweiterte suche. TODO what to do?
        # 
        # handle both STRING and LONGSTRING as StringRec.  115``
        elif recval["laborvalue_type"] == "STRING" or recval["laborvalue_type"] == "LONGSTRING":
            out = StringRec(method=finding["method"], labval=recval["laborvalue_code"], rec=recval["stringvalue"])
        # take DATE values from datevalue and LONGDATE values from
        # datevalueprecision, respectively.  116``
        elif recval["laborvalue_type"] == "DATE":
            out = DateRec(method=finding["method"], labval=recval["laborvalue_code"], rec=recval["datevalue"])
        elif recval["laborvalue_type"] == "LONGDATE":
            out = DateRec(method=finding["method"], labval=recval["laborvalue_code"], rec=recval["datevalueprecision"])
        # for catalog, query the actual catalog entries. go from recordedvalue
        # to catalogentry via the recordedval_catentry table.  117``
        elif recval["laborvalue_type"] == "CATALOG":
            # get the catalog code
            query = """select catalog.code as 'catalog_code' from centraxx_catalog as catalog
            where catalog.oid = ?""" # do this once for all catalogs on startup or finding() call?
            res = self.db.qfad(query, recval['laborvalue_catalog_oid'])
            catalog_code = res[0]["catalog_code"]
            # get the catalog entries
            query = """select catalogentry.code as 'catalogentry_code' from centraxx_recordedvalue as recordedvalue
            join centraxx_recordedval_catentry as recordedval_catentry on recordedval_catentry.recordedvalue_oid = recordedvalue.oid
            join centraxx_catalogentry as catalogentry on catalogentry.oid = recordedval_catentry.catalogentry_oid
            where recordedvalue.oid = ?"""
            res = self.db.qfad(query, recval['oid'])
            # get the catalogentry names if they are not already loaded, and cache them. maybe it's faster to load the whole names map once instead of joining them in each time?
            if names is True and self.names_catalogentry is None:
                self.names_catalogentry = self.name(table="catalogentry")
            entries = []
            value_name = {}
            for r in res:
                code = r["catalogentry_code"]
                entries.append(code)
                if names is True:
                    value_name[code] = self.names_catalogentry[code]
            
            out = CatalogRec(method=finding["method"], labval=recval["laborvalue_code"], catalog=catalog_code, rec=entries, rec_name=value_name)
        # for enumeration and option group, query the actual usage entries. go from recordedvalue
        # to usageentry via the recordedval_usageentry table.
        # 
        # for now, use a MultiRec type, that is the same as CatalogRec type,
        # except without catalog code?  118``
        elif recval["laborvalue_type"] == "ENUMERATION" or recval["laborvalue_type"] == "OPTIONGROUP":
            query = """select usageentry.code as 'usageentry_code' from centraxx_recordedvalue as recordedvalue
            join centraxx_recordedval_usagentry as recordedval_usagentry on recordedval_usagentry.recordedvalue_oid = recordedvalue.oid
            join centraxx_usageentry as usageentry on usageentry.oid = recordedval_usagentry.usageentry_oid
            where recordedvalue.oid = ?"""
            res = self.db.qfad(query, recval['oid'])

            # get the usageentry names if they are not already loaded, and cache them. maybe it's faster to load the whole names map once instead of joining them in each time?
            if names is True and self.names_usageentry is None:
                self.names_usageentry = self.name(table="usageentry")

            entries = []
            value_name = {}
            for r in res:
                code = r["usageentry_code"]
                entries.append(code)
                if names is True:
                    value_name[code] = self.names_usageentry[code]
            
            out = MultiRec(method=finding["method"], labval=recval["laborvalue_code"], rec=entries, rec_name=value_name)
        # else throw an exception.  119``
        else:
            raise Exception(f"no record class for laborvalue of type {recval['laborvalue_type']}")
        # return.  120``
        return out
        #``113
    #``112
    # _query sticks together the query from the various lists and parameters and returns the result: (res, bydict, missinglst, by)
    # 
    # tablename is the name of the table without 'centraxx_' prefix.
    # 
    # list is the dict of query-parameter lists that were passed to the
    # calling function.  383``/_query:
    def _query(self, tablename:str=None, lists:dict=None, idc:dict=None, sampleids:list=None, patientids:list=None, pidc:str=None, files:dict=None, verbose:list=None, verbose_all:bool=False, like:list=None, missing:bool=False, order_by:str=None, top:int=None, by:str=None, print_query:bool=None, vaa:list=None, jselect:dict=None, lselect:dict=None, joins:dict=None):
        # init arrays.  don't set them as mutable default parameters
        # (e.g. verbose=[]), cause they would then retain their values across
        # method calls. see
        # https://www.geeksforgeeks.org/python/use-mutable-default-value-as-an-argument-in-python/  384``_:
       	if lists is None:
       	    lists = {}
               if idc is None:
                   idc = {}
       	if sampleids is None:
       	    sampleids = []
       	if patientids is None:
       	    patientids = []
               if files is None:
                   files = {}
               if verbose is None:
                   verbose = []
               if like is None:
                   like = []
       	if vaa is None:
       	    vaa = []
       	if jselect is None:
       	    jselect = {}
       	if lselect is None:
       	    lselect = {}
       	if joins is None:
       	    joins = {}
        # on verbose_all include all constants in the verbose array.  385`` #py
               if verbose_all:
                   verbose = vaa
        # replace traction constants patientid and sampleid with their idc.  386``
               verbose = self._concrete_idcs(verbose, pidc=pidc)
               files = self._concrete_idcs_dict(files, pidc=pidc)        
        # make sure that the verbose array only contains keys from the verbose-all-array.  387``
               if not _checkverbose(verbose, vaa): 
                   raise Exception(f"verbose keys {verbose} need to be in {vaa}.")
        # resolve the sampleid and patientid shortcut by putting their lists
        # into the idc dict.  388``
               _dextend(idc, self.sidc(), sampleids)
               _dextend(idc, self.pidc(pidc), patientids)
        # now idc holds all parameters queryable as idcontainers.
        # 
        # read the lists from files and put them together with the lists passed.  389``
               alllists = self._makealllists(lists, idc, files)
        # put all lists longer than 50 items into temporary tables.
        # 
        # nontable holds the remaining lists, table holds the names of the created
        # temporary tables, each distinguishing nonidc and idc.  390``
               (nontable, table) = self._makemove(alllists, 50)
        # put together the select string.  391``
               selectstr = self._selectstr(jselects, lselects, nontable, table, verbose)
        # put together the join string for the query.  392`` #py
               joinstr = self._joinstr(joins, nontable, table, verbose, pidc=pidc)
        # get the where string.  393`` #py
               (wherestr, whereargs) = self._where(nontable, table, like=like)
        # get the top string.  394``
               topstr = self._top(top)
        # stick together the query, get the result and return it.  395`` #py
               query = f"select {topstr} {selectstr} from centraxx_{tablename} {tablename} \n{joinstr}"
               if wherestr.strip() != "":
                   query += f"\nwhere {wherestr}"
               if order_by is not None:
                   query += " " + self._order_by(order_by)
                   
               if print_query:
                  print(query)
                  print(whereargs)
       
               res = self.db.qfad(query, whereargs)
        # clear the temporary tables.  396``
               self._cleartt(table["nonidc"])
               self._cleartt(table["idc"])
        # if missing is passed, and no by, and there's only one query
        # parameter, set by to that query parameter.  397``
               if missing and by is None:
                   by = _byifone(alllists)
        # if a by is passed, return the samples keyed by that value.
        # 
        # if missing is True alongside by, only return what's missing.  398``
               bydict, missinglst = _prepby(alllists, by, missing)
        # if missing was passed, delete the by-key from the missing list.  399``
               for r in res:
                   if by is not None and missing:
                       _updatemissing(missinglst, by, r)
        # return the query result, the bydict, the missing list and `by`. return
        # `by` cause it might have been changed above.  400``
               return (res, bydict, missinglst, by)
        #``384
    #``383
    
    # _fill_in_primary adds a reference to the given sample's primary sample,
    # if there is one.  230``/_fill_in_primary:
    def _fill_in_primary(self, sample:Sample):
        # get the parent oids.   231.``
        poids = []
        self._get_parents(sample.id("oid"), poids)
        # the first parent is the primary, take it.  232``
        if len(poids) > 0:
            primary_oid = poids[0]
            # query the primary to get its sampleid, reference it from the sample as
            # Idable.  233.``
            primary = self.sample(oids=[primary_oid])[0]
            sample.primary = Idable(ids=primary.ids, mainidc=primary.mainidc)
            #``233
        #``231
    #``230

    # lists and tmp tables
    # _makemove makes temporary tables from the given lists that are longer
    # than cutoff. it returns one dict holding the remaining, non-table
    # lists and one dict holding the table names of the created temporary
    # tables, both return dicts distinguishing between nonidc and idc.  290``/_makemove:
    def _makemove(self, alllists:dict, cutoff:int): # -> (dict, dict)#bm
        # move large argument lists into the temporary table dicts.  291.``
        nontables = {}
        (nontables["nonidc"], ttlists) = _tablelists(alllists["nonidc"], cutoff)
        (nontables["idc"], ttidclists) = _tablelists(alllists["idc"], cutoff)        
        # create the temporary tables.  292``
        tables = {}
        tables["nonidc"] = self._maketables(ttlists, "tmp_")
        tables["idc"] = self._maketables(ttidclists, "tmp_idc_")
        # return nontable lists and table names.  293``
        return (nontables, tables) 
        #``291
    #``290
    # _makealllists reads the lists from files, puts them together with the
    # lists passed as lists, and returns them all in one dict,
    # keeping nonidc and idc seperate.  294``/_makealllists:
    def _makealllists(self, lists, idc, files):
        # make the return dict with one chapter for non idc and one for idc.  295``_:
        alllists = {}
        alllists["nonidc"] = {}
        alllists["idc"] = {}
        # put in the lists.  296``
        for key, lst in lists.items():
            if lst is not None:
                alllists["nonidc"][key] = lst
        for key, lst in idc.items():
            alllists["idc"][key] = lst
        # read the lists from files and split them into non-idc and idc.  297``
        (ff, ffidc) = self._readfiles(files)
        # put in the lists from files.  298``
        for key, lst in ff.items():
            if key not in alllists["nonidc"]:
                alllists["nonidc"][key] = []
            alllists["nonidc"][key].extend(lst)
        for key, lst in ffidc.items():
            if key not in alllists["idc"]:
                alllists["idc"][key] = []
            alllists["idc"][key].extend(lst)
        # return.  299``
        return alllists
        #``295
    #``294
    # _readfiles reads the files from which tmp tables should be build into
    # lists.  it returns a tuple of one dict holding the non-idc lists and
    # one dict holding the idc lists.  300``/_readfiles:
    def _readfiles(self, files, sidc:str=None, pidc:str=None):
        # if no files, return empty.  301.``
        if files is None:
            return ({}, {})
        # read the files.  302``
        both = {}
        for key, filepath in files.items():
            lst = []
            with open(filepath, "r") as f:
                for line in f:
                    lst.append(line.rstrip())
            both[key] = lst
        # split into non-idc and idc dicts.  303``
        nonidc = {}
        idc = {}
        for key, _ in both.items():
            if key == sampleid:
                idc[self.sidc()] = both[key]
            elif key == patientid:
                idc[self.pidc(pidc)] = both[key]
            elif self._is_idc(key, sidc=sidc, pidc=pidc):
                idc[key] = both[key]
            else:
                nonidc[key] = both[key]
        # return non-idc and idc as tuple.  304``
        return (nonidc, idc)
        #``301
    #``300
    # _maketables makes temporary tables for each key in the given dict,
    # inserts the list of values belonging to that key, and returns a dict
    # the table names by key.  305``/_maketables:
    def _maketables(self, d:dict, prefix:str=""): #-> dict
        # make the temporary tables, insert the values and remember the table names.  306.``
        out = {}
        for key, lst in d.items():
            #name = "tempdb..#" + prefix + key
            name = "#" + prefix + key
            if not isidentifier(name):
                raise Exception(f"name {name} needs to be a valid sql identifier.")
            self.db.query(f"create table {name} (stdin varchar (255) )")
            for e in lst:
                self.db.query(f"insert into {name} values (?)", e)
            out[key] = name
        # return the created table names.  307``
        return out
        #``306
    #``305
    # _cleartt drops the temporary tables tables that were created with
    # _maketables after the query using their data was run. it takes the
    # dict of filetables that was returned by _maketables.
    # 
    # they would be dropped automatically after the session, but in case
    # several queries are sent in one session, maybe it's better to drop
    # them after each query.  288``/_cleartt:
    def _cleartt(self, tmptables:dict):
        # go over the tables and drop them.  289.``
        for key, tablename in tmptables.items():
            if not isidentifier(tablename):
                raise Exception("table name {tablename} needs to be a valid sql identifier.")
            q = f"drop table tempdb..{tablename}"
            #print(q)
            self.db.query(q)  # todo is kairos_spring used again after this?
        #``289
    #``288

    # settings / idc related
    # sidc returns the main idc code by which samples are referenced as
    # specified in the settings.   210``/sidc:
    def sidc(self) -> str:
        if self.db.target not in self.settings['sampleid']:
            raise Exception(f"please specify the main sample idcontainer for {self.db.target} in settings.yaml")
        return self.settings['sampleid'][self.db.target]
    #``210
    # sidc returns the main idc code by which patients are referenced as
    # specified in the settings. if a pidc argument is given, it is returned
    # instead of the settings' pidc.  211``/pidc:
    def pidc(self, pidc:str|None=None) -> str:
        if pidc is not None:
            return pidc
        if self.db.target not in self.settings['patientid']:
            raise Exception(f"please specify the main patient idcontainer for {self.db.target} in settings.yaml")
        return self.settings['patientid'][self.db.target]
    #``211
    # _sidcs returns the idcs from settings that are specific for sample.
    # 
    # rather make this a section in conf? but maybe not, keep the conf simple and let the program sort out pidc and sidc.
    # 
    # # idc holds idcontainer codes that should be queryable as command line flags 
    # idc:
    # sample:
    # - extsampleid
    # - modul
    # - tier
    # patient:
    #     - mpi  213``/_sidcs:
    def _sidcs(self) -> list:
        # the settings can list idc codes that are not applicable to this
        # target, skip them.  214.``
        out = []
        for idc in self.settings["idc"]:
            if idc in self._idckind and self._idckind[idc] == "SAMPLE" and not idc in out:
                out.append(idc)
        # include the main sample idcontainer. it could already be in out.  215``
        sidc = self.sidc()
        if sidc not in out:
            out.append(self.sidc())
        return out
        #``214
    #``213
    # _pidcs returns the idcs from settings that are specific for
    # sample.  216``/_pidcs:
    def _pidcs(self, pidc:str=None) -> list:
        # the settings can list idc codes that are not applicable to this target, skip them.  217.``
        out = []
        for idc in self.settings["idc"]:
            if idc in self._idckind and self._idckind[idc] == "PATIENT" and idc not in out:
                out.append(idc)
        # include the main patient idcontainer. it could be already in the array.  218``
        pidc = self.pidc(pidc)
        if pidc not in out:
            out.append(pidc)
        return out
        #``217
    #``216
    # _concrete_idcs replaces traction constants patientid and sampleid with their
    # respective idc for array.  219``/_concrete_idcs:
    def _concrete_idcs(self, verbose, pidc=None):
        # fill in the concrete idcs when encountering patientid or sampleid.  220.``
        out = []
        for verb in verbose:
            if verb == patientid:
                out.append(self.pidc(pidc))
            elif verb == sampleid:
                out.append(self.sidc())
            else:
                out.append(verb)
        return out
        #``220
    #``219
    # _concrete_idcs_dict replaces traction constants patientid and sampleid with their
    # respective idc for dict.  221``/_concrete_idcs_dict:
    def _concrete_idcs_dict(self, d, pidc=None):
        # fill in the concrete idcs when encountering patientid or sampleid.  222.``
        out = {}
        for key, val in d.items():
            #print("key: " + key)
            if key == patientid:
                out[self.pidc(pidc)] = val
            elif key == sampleid:
                #print("here: " + self.sidc())
                out[self.sidc()] = val
            else:
                out[key] = val
        return out
        #``222
    #``221
    # cxx gives the centraxx version for db target from settings.  212``/cxx:
    def cxx(self) -> str:
        v = dig(self.settings, f"cxx/{self.db.target}")
        if v is None:
            return None
        return str(v)
    #``212

    # _is_idc says whether a string is an idcontainer mentioned in
    # settings.yaml or passed as sidc or pidc argument.  223``/_is_idc:
    def _is_idc(self, key:str, sidc:str=None, pidc:str=None): # -> bool
        return key in self.settings["idc"] or key == self.sidc() or key == self.pidc(pidc)
    #``223
    # _splitidc splits an array of tr constants and idc keys into non-idc
    # (tr constants) and idc keys.  224``/_splitidc:
    def _splitidc(self, arr:list, pidc:str=None): # -> (list, list)
        # is the verbose key idc?  225.``
        nonidc = []
        idc = []
        for k in arr:
            if self._is_idc(k, pidc=pidc):
                idc.append(k)
            else:
                nonidc.append(k)
        return (nonidc, idc)
        #``225
    #``224
    # _collkeys collects the keys join needs from lists, temporary tables
    # and verbose. preserves order and ensures uniqueness.  226``/_collkeys:
    def _collkeys(self, lists:dict=None, tmptables:dict=None, verbose:list=None):
        # init None params.  227.``
        if lists is None:
            lists = {}
        if tmptables is None:
            tmptables = {}
        if verbose is None:
            verbose = []
        # add the keys of lists and temporary tables.  228``
        out = []
        for k in lists.keys():
            if lists[k] is not None and k not in out:
                out.append(k)
        for k in tmptables.keys():
            if tmptables[k] is not None and k not in out:
                out.append(k)
        # append the items of the verbose list and return.  229``
        for k in verbose:
            if k not in out:
                out.append(k)
        return out
        #``227
    #``226
    
#``1
