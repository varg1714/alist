import{q as o,r as e,s as F,t as n,v as r,w as m,x as E,y as c,z as d,A as B,C as u,D as I,E as k}from"./index.e26d2df3.js";import{f as p,g as x,i as D}from"./index.577e9bef.js";import{bG as f,bH as A}from"./index.9265efac.js";import{I as g}from"./Layout.f8eadd64.js";var s=(a=>(a[a.UNKNOWN=0]="UNKNOWN",a[a.FOLDER=1]="FOLDER",a[a.VIDEO=2]="VIDEO",a[a.AUDIO=3]="AUDIO",a[a.TEXT=4]="TEXT",a[a.IMAGE=5]="IMAGE",a))(s||{});const M={"dmg,ipa,plist,tipa":m,"exe,msi":E,"zip,gz,rar,7z,tar,jar,xz":c,apk:g,db:p,md:d,epub:x,iso:D,m3u8:r,"doc,docx":B,"xls,xlsx":u,"ppt,pptx":I,pdf:k},N=(a,i)=>{if(a!==s.FOLDER){for(const[l,t]of Object.entries(M))if(l.split(",").includes(i.toLowerCase()))return t}switch(a){case s.FOLDER:return A;case s.VIDEO:return r;case s.AUDIO:return n;case s.TEXT:return F;case s.IMAGE:return e;default:return o}},G=a=>N(a.type,f(a.name));export{s as O,G as g};
