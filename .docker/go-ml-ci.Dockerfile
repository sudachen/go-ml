FROM sudachen/go1137-ci:latest
LABEL maintainer="Alexey Sudachen <alexey@sudachen.name>"

USER root
RUN curl -L https://github.com/sudachen/mxnet/releases/download/1.5.0-mkldnn-static/libmxnet_cpu_lin64.lzma -o /tmp/mxnet.lzma \
 && mkdir -p /opt/mxnet/lib \
 && xz -d -c /tmp/mxnet.lzma > /opt/mxnet/lib/libmxnet.so
RUN curl -L https://github.com/sudachen/xgboost/releases/download/custom/libxgboost_cpu_lin64.lzma -o /tmp/xgboost.lzma \
 && mkdir -p /opt/xgboost/lib \
 && xz -d -c /tmp/xgboost.lzma > /opt/xgboost/lib/libxgboost.so

USER $USER
CMD ["/bin/sh"]
