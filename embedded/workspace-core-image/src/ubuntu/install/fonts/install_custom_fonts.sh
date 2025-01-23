#!/usr/bin/env bash
### every exit != 0 fails the script
set -e

LOCALES="aa_DJ aa_ER aa_ET af_ZA am_ET an_ES ar_AE ar_BH ar_DZ ar_EG ar_IN ar_IQ ar_JO ar_KW ar_LB ar_LY ar_MA ar_OM ar_QA ar_SA ar_SD ar_SY ar_TN ar_YE as_IN ast_ES ayc_PE az_AZ be_BY bem_ZM ber_DZ ber_MA bg_BG bho_IN bn_BD bn_IN bo_CN bo_IN br_FR brx_IN bs_BA byn_ER ca_AD ca_ES ca_FR ca_IT crh_UA csb_PL cs_CZ cv_RU cy_GB da_DK de_AT de_BE de_CH de_DE de_LU doi_IN dv_MV dz_BT el_CY el_GR en_AG en_AU en_BW en_CA en_DK en_GB en_HK en_IE en_IN en_NG en_NZ en_PH en_SG en_US en_ZA en_ZM en_ZW es_AR es_BO es_CL es_CO es_CR es_CU es_DO es_EC es_ES es_GT es_HN es_MX es_NI es_PA es_PE es_PR es_PY es_SV es_US es_UY es_VE et_EE eu_ES fa_IR ff_SN fi_FI fil_PH fo_FO fr_BE fr_CA fr_CH fr_FR fr_LU fur_IT fy_DE fy_NL ga_IE gd_GB gez_ER gez_ET gl_ES gu_IN gv_GB ha_NG he_IL hi_IN hne_IN hr_HR hsb_DE ht_HT hu_HU hy_AM ia_FR id_ID ig_NG ik_CA is_IS it_CH it_IT iu_CA ja_JP ka_GE kk_KZ kl_GL km_KH kn_IN kok_IN ko_KR ks_IN ku_TR kw_GB ky_KG lb_LU lg_UG li_BE lij_IT li_NL lo_LA lt_LT lv_LV mag_IN mai_IN mg_MG mhr_RU mi_NZ mk_MK ml_IN mni_IN mn_MN mr_IN ms_MY mt_MT my_MM nb_NO nds_DE nds_NL ne_NP nhn_MX niu_NU niu_NZ nl_AW nl_BE nl_NL nn_NO nr_ZA nso_ZA oc_FR om_ET om_KE or_IN os_RU pa_IN pa_PK pl_PL ps_AF pt_BR pt_PT ro_RO ru_RU ru_UA rw_RW sa_IN sat_IN sc_IT sd_IN se_NO shs_CA sid_ET si_LK sk_SK sl_SI so_DJ so_ET so_KE so_SO sq_AL sq_MK sr_ME sr_RS ss_ZA st_ZA sv_FI sv_SE sw_KE sw_TZ szl_PL ta_IN ta_LK te_IN tg_TJ th_TH ti_ER ti_ET tig_ER tk_TM tl_PH tn_ZA tr_CY tr_TR ts_ZA tt_RU ug_CN uk_UA unm_US ur_IN ur_PK uz_UZ ve_ZA vi_VN wa_BE wae_CH wal_ET wo_SN xh_ZA yi_US yo_NG yue_HK zh_CN zh_HK zh_SG zh_TW zu_ZA"

echo "Installing fonts and languages"
  zypper addrepo -G \
    https://download.opensuse.org/repositories/M17N:/fonts/15.6/ fonts-x86_64
  zypper install -ny \
    glibc-i18ndata \
    glibc-locale \
    google-noto-coloremoji-fonts \
    google-noto-sans-cjk-fonts \
    noto-sans-fonts

  for LOCALE in ${LOCALES}; do
    echo "Generating Locale for ${LOCALE}"
    localedef -i ${LOCALE} -f UTF-8 ${LOCALE}.UTF-8
  done