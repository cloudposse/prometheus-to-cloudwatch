FROM ghcr.io/fluxcd/helm-controller:v0.24.0

RUN cd /usr/share/ca-certificates/ && curl http://certinfo.roche.com/rootcerts/Roche%20Root%20CA%201.cer -o RocheRootCA1.cer && curl http://certinfo.roche.com/rootcerts/RocheEnterpriseCA1.cer -o RocheEnterpriseCA1.cer && curl http://certinfo.roche.com/rootcerts/RocheEnterpriseCA2.cer -o RocheEnterpriseCA2.cer && curl http://certinfo.roche.com/rootcerts/Roche%20Root%20CA%201%20-%20G2.crt -o RocheRootCA1-G2.crt && curl http://certinfo.roche.com/rootcerts/Roche%20Enterprise%20CA%201%20-%20G2.crt -o RocheEnterpriseCA1-G2.crt && curl http://certinfo.roche.com/rootcerts/Roche%20G3%20Root%20CA.crt -o RocheG3RootCA.crt && curl http://certinfo.roche.com/rootcerts/Roche%20G3%20Issuing%20CA%201.crt -o RocheG3IssuingCA1.crt && curl http://certinfo.roche.com/rootcerts/Roche%20G3%20Issuing%20CA%202.crt -o RocheG3IssuingCA2.crt && curl http://certinfo.roche.com/rootcerts/Roche%20G3%20Issuing%20CA%203.crt -o RocheG3IssuingCA3.crt && curl http://certinfo.roche.com/rootcerts/Roche%20G3%20Issuing%20CA%204.crt -o RocheG3IssuingCA4.crt && update-ca-certificates

USER 65534:65534

ENTRYPOINT [ "/sbin/tini", "--", "helm-controller" ]
