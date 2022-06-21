#!/bin/sh
export ALTAIR_HOME=/opt/compose2021.2/altair/Compose2021.2
export HW_ROOTDIR=/opt/compose2021.2/altair/Compose2021.2
export HW_UNITY_ROOTDIR=/opt/compose2021.2/altair/Compose2021.2/hwx
export HW_PLATFORM=linux64
export PLATFORM=linux64
export ALTAIR_PROD_ARCH=linux64
export PATH=$PATH:/opt/compose2021.2/altair/Compose2021.2/hwx/bin/linux64:/opt/compose2021.2/altair/Compose2021.2/hw/bin/linux64
export LD_LIBRARY_PATH=/opt/compose2021.2/altair/Compose2021.2/hwx/bin/linux64:/opt/compose2021.2/altair/Compose2021.2/hw/bin/linux64
export HW_EDITION=business

JAVA_OPTS="$JAVA_OPTS -Dconfig_dir=\"$PBSWORKS_HOME\""
JAVA_OPTS="$JAVA_OPTS -DCONFIG_HOME=\"$PBSWORKS_HOME\""
JAVA_OPTS="$JAVA_OPTS -DHWE_INSTALLATION_DIRECTORY=\"$PBSWORKS_EXEC\""
JAVA_OPTS="$JAVA_OPTS -DHWE_HOME_DIRECTORY=\"$PBSWORKS_HOME\""
JAVA_OPTS="$JAVA_OPTS -DHWE_RM_LIB_HOME=\"$PBSWORKS_EXEC/resultmanager/binaries/rvs_lib\""
JAVA_OPTS="$JAVA_OPTS -DHWE_RM_LIB_HOME_DIR=\"$PBSWORKS_HOME/config/resultmanager\""
JAVA_OPTS="$JAVA_OPTS -DHWE_REST_DOC_CONFIG_DIR=\"$PBSWORKS_EXEC/resultmanager/binaries/rvs_lib/resources/rest_doc_gen\""
JAVA_OPTS="$JAVA_OPTS -Xmx${__services_resultmanager_max_memory}"
JAVA_OPTS="$JAVA_OPTS -Dlog4j2.configurationFile=\"$PBSWORKS_HOME/config/resultmanager/log4j2.properties\""
JAVA_OPTS="$JAVA_OPTS -Djavax.xml.bind.context.factory=org.eclipse.persistence.jaxb.JAXBContextFactory"	

export JAVA_OPTS