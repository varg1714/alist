import{q as e,r as o,s as c,t as n,v as r,w as F,x as m,y as E,z as d,A as u,C as B,D as I,E as p}from"./index.f3072a2b.js";import{f as x,g as f,i as k}from"./index.abc04cae.js";import{al as A,bG as D,bH as v}from"./index.c7be97ab.js";var s=(a=>(a[a.UNKNOWN=0]="UNKNOWN",a[a.FOLDER=1]="FOLDER",a[a.VIDEO=2]="VIDEO",a[a.AUDIO=3]="AUDIO",a[a.TEXT=4]="TEXT",a[a.IMAGE=5]="IMAGE",a))(s||{});function z(a){return A({a:{viewBox:"0 0 16 16"},c:'<path fill="currentColor" d="M14 6c-.55 0-1 .45-1 1v4c0 .55.45 1 1 1s1-.45 1-1V7c0-.55-.45-1-1-1zM2 6c-.55 0-1 .45-1 1v4c0 .55.45 1 1 1s1-.45 1-1V7c0-.55-.45-1-1-1zm1.5 5.5A1.5 1.5 0 005 13v2c0 .55.45 1 1 1s1-.45 1-1v-2h2v2c0 .55.45 1 1 1s1-.45 1-1v-2a1.5 1.5 0 001.5-1.5V6h-9v5.5zM12.472 5a4.5 4.5 0 00-2.025-3.276l.5-1.001a.5.5 0 00-.895-.447L9.55 1.28l-.13-.052a4.504 4.504 0 00-2.84 0l-.13.052L5.948.276a.5.5 0 00-.895.447l.5 1.001A4.499 4.499 0 003.528 5v.5H12.5V5h-.028zM6.5 4a.5.5 0 01-.001-1h.002A.5.5 0 016.5 4zm3 0a.5.5 0 01-.001-1h.003a.5.5 0 01-.001 1z"/>'},a)}const M={"dmg,ipa,plist,tipa":F,"exe,msi":m,"zip,gz,rar,7z,tar,jar,xz":E,apk:z,db:x,md:d,epub:f,iso:k,m3u8:r,"doc,docx":u,"xls,xlsx":B,"ppt,pptx":I,pdf:p},g=(a,l)=>{if(a!==s.FOLDER){for(const[i,t]of Object.entries(M))if(i.split(",").includes(l.toLowerCase()))return t}switch(a){case s.FOLDER:return v;case s.VIDEO:return r;case s.AUDIO:return n;case s.TEXT:return c;case s.IMAGE:return o;default:return e}},w=a=>g(a.type,D(a.name));export{s as O,w as g};
